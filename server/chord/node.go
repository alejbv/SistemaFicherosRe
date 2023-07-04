package chord

import (
	"context"
	"errors"
	"math/big"
	"net"
	"sync"

	"github.com/alejbv/SistemaFicherosRe/chord"
	"github.com/alejbv/SistemaFicherosRe/server/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Node struct {
	*chord.Node // Nodo real.

	predecessor *chord.Node        // Predecesor de este nodo en el anillo.
	predLock    sync.RWMutex       // Bloquea el predecesor para lectura o escritura.
	successors  *Queue[chord.Node] // Cola de sucesores de este nodo en el anillo.
	sucLock     sync.RWMutex       // Bloquea la cola de sucesores para lectura o escritura.

	fingerTable FingerTable  // FingerTable de este nodo.
	fingerLock  sync.RWMutex // Bloquea la FingerTable para lectura o escritura.

	RPC    RemoteServices // Capa de transporte para este nodo(implementa la parte del cliente del chord).
	config *Configuration // Configuraciones generales.

	// Los metodos con el dictionario se deben modificar
	dictionary storage.Storage // Diccionario de almacenamiento de este nodo.
	dictLock   sync.RWMutex    // Bloquea el diccionario para lectura o escritura.

	server   *grpc.Server     // Nodo servidor.
	sock     *net.TCPListener // Socket para escuchar conexiones del nodo servidor.
	shutdown chan struct{}    // Determina si el nodo esta actualemente en ejecucion.

	chord.UnimplementedChordServer
}

// NewNode crea y devuelve un nuevo nodo.
func NewNode(port string, configuration *Configuration, transport RemoteServices, dictionary storage.Storage) (*Node, error) {
	// Su la configuracion es vacia devuelve error.
	if configuration == nil {
		log.Error("Error creando el nodo: la configuracion no puede ser vacia.")
		return nil, errors.New("error creando el nodo: la configuracion no puede ser vacia")
	}

	// Crea un nuevo nodo con la ID obtenida y la misma dirección.
	innerNode := &chord.Node{ID: big.NewInt(0).Bytes(), IP: "0.0.0.0", Port: port}

	// Crea la instancia del nodo.
	node := &Node{Node: innerNode,
		predecessor: nil,
		successors:  nil,
		fingerTable: nil,
		RPC:         transport,
		config:      configuration,
		dictionary:  dictionary,
		server:      nil,
		shutdown:    nil}

	// Devuelve el nodo.
	return node, nil
}

// DefaultNode crea y devuelve un nuevo nodo con una configuracion por defecto.
func DefaultNode(port string, dictionary storage.Storage) (*Node, error) {
	conf := DefaultConfig() // Crea una configuracion por defecto.
	transport := NewGRPCServices(conf)
	//path := "/usr/app/server/data/" // Crea un objeto RPC por defecto  para interactuar con la capa de transporte.
	//dictionary := storage.Storage(conf.Hash, path) // Crea un diccionario por defecto.

	// Devuelve el nodo creado.
	return NewNode(port, conf, transport, dictionary)
}

//Ok
// GetPredecessor devuelve el nodo que se cree que es el actual predecesor.
func (node *Node) GetPredecessor(ctx context.Context, req *chord.GetPredecessorRequest) (*chord.GetPredecessorResponse, error) {
	log.Trace("Obteniendo el predecesor del nodo.")

	// Bloquea el predecesor para leerlo, se desbloquea el terminar.
	node.predLock.RLock()
	pred := node.predecessor
	node.predLock.RUnlock()

	// Se crea el mensaje respuesta que contiene al nodo predecesor
	res := &chord.GetPredecessorResponse{
		Predecessor: pred,
	}
	// Devuelve el predecesor de este nodo.
	return res, nil
}

//Ok
// GetSuccessor regresa el nodo que se cree que es el sucesor actual del nodo.
func (node *Node) GetSuccessor(ctx context.Context, req *chord.GetSuccessorRequest) (*chord.GetSuccessorResponse, error) {
	log.Trace("Obteniendo el sucesor del nodo.")

	// Bloquea el sucesor para leer de el, se desbloquea el terminar
	node.sucLock.RLock()
	suc := node.successors.Beg()
	node.sucLock.RUnlock()

	// Devuelve el sucesor de este nodo
	res := &chord.GetSuccessorResponse{Successor: suc}
	return res, nil
}

//Ok
// SetPredecessor establece el predecesor de este nodo.
func (node *Node) SetPredecessor(ctx context.Context, req *chord.SetPredecessorRequest) (*chord.SetPredecessorResponse, error) {
	log.Trace("Fijando el sucesor del nodo.")

	candidate := req.GetPredecessor()
	log.Tracef("Estableciendo el nodo predecesor a %s.", candidate.IP)

	// Si el nuevo predecesor no es este nodo, se actualiza el nuevo predecesor
	if !Equals(candidate.ID, node.ID) {
		// Bloquea el predecesor para lectura y escritura, se desbloquea el finalizar
		node.predLock.Lock()
		old := node.predecessor
		node.predecessor = candidate
		node.predLock.Unlock()
		// Si existia un anterior predecesor se absorven sus claves
		go node.AbsorbPredecessorKeys(old)
	} else {
		log.Trace("El candidato a predecesor es el mismo nodo. No es necesario actualizar.")
	}

	return &chord.SetPredecessorResponse{}, nil
}

//Ok
// SetSuccessor establece el sucesor de este nodo.
func (node *Node) SetSuccessor(ctx context.Context, req *chord.SetSuccessorRequest) (*chord.SetSuccessorResponse, error) {

	candidate := req.GetSuccessor()
	log.Tracef("Estableciendo el nodo sucesor a %s.", candidate.IP)

	// Si el nuevo sucesor es distinto al actual nodo, se actualiza
	if !Equals(candidate.ID, node.ID) {
		//Bloquea el sucesor para escribir en el, se desbloquea al terminar
		node.sucLock.Lock()
		node.successors.PushBeg(candidate)
		node.sucLock.Unlock()
		// Actualiza este nuevo sucesor con las llaves de este nodo
		go node.UpdateSuccessorKeys()
	} else {
		log.Trace("Candidato a sucesor es el mismo nodo. No hay necesidad de actualizar.")
	}

	return &chord.SetSuccessorResponse{}, nil
}

//Ok
// FindSuccessor busca el nodo sucesor de esta ID.
func (node *Node) FindSuccessor(ctx context.Context, req *chord.FindSuccesorRequest) (*chord.FindSuccesorResponse, error) {
	// Busca el sucesor de esta ID.
	node_id := req.ID
	new_node, err := node.FindIDSuccessor(node_id)
	if err != nil {
		return nil, err
	}
	succesor := chord.Node{ID: new_node.ID, IP: new_node.IP, Port: new_node.Port}
	res := &chord.FindSuccesorResponse{Succesor: &succesor}
	return res, nil
}

//OK
// Notify notifica a este nodo que es posible que tenga un nuevo predecesor.
func (node *Node) Notify(ctx context.Context, req *chord.NotifyRequest) (*chord.NotifyResponse, error) {
	log.Info("Comprobando la notificacion de predecesor.")

	candidate := req.Notify

	// Bloquea el predecesor para leer de el, lo desbloquea al terminar.
	node.predLock.RLock()
	pred := node.predecessor
	node.predLock.RUnlock()
	/*

		Si este nodo no tiene predecesor o el candidato a predecesor esta más cerca a este
		nodo que su actual predecesor, se actualiza el predecesor con el candidato
	*/
	if Equals(pred.ID, node.ID) || Between(candidate.ID, pred.ID, node.ID) {
		log.Infof("Predecesor actualizado al nodo en %s.", candidate.IP)

		// Bloquea el predecesor para escribir en el, lo desbloquea al terminar.
		node.predLock.Lock()
		node.predecessor = candidate
		node.predLock.Unlock()

		//Actualiza el nuevo predecesor con la correspondiente clave.
		go node.UpdatePredecessorKeys(pred)
	}

	return &chord.NotifyResponse{}, nil
}

//Ok
//Comprueba si el nodo esta vivo.
func (node *Node) Check(ctx context.Context, req *chord.CheckRequest) (*chord.CheckResponse, error) {
	return &chord.CheckResponse{}, nil
}

//Ok
// Añade un archivo al almacenamiento local del nodo correspondiente
func (node *Node) Set(ctx context.Context, req *chord.SetRequest) (*chord.SetResponse, error) {
	log.Infof("Set: key=%s.", req.Ident)
	// Obtenemos el sucesor del nodo actual
	successor, err := node.LocateKey(req.Ident)
	if err != nil {
		log.Error("Error getting key." + err.Error())
		return &chord.SetResponse{}, err
	}

	// Si el sucesor es igual al nodo actual o es una replicas, entonces tenemos el valor
	if Equals(successor.ID, node.ID) || req.Replication {

		//Se bloque el diccionario para escribir en el
		node.dictLock.Lock()
		err := node.dictionary.Set(req.Ident, req.File)
		//Se desbloquea al terminar de escribir en el
		node.dictLock.Unlock()
		if err != nil {
			log.Error("Error getting value." + err.Error())
			return &chord.SetResponse{}, err
		}
		if req.Replication {
			log.Info("Resolviendo la request Set localmente (replicacion).")
		} else {
			// Se bloquea el sucesor para leer de el, se desbloquea al terminar
			node.sucLock.RLock()
			suc := node.successors.Beg()
			node.sucLock.RUnlock()
			// Si el sucesor es distinto a este nodo, se replica la request
			if !Equals(suc.ID, node.ID) {
				go func() {
					req.Replication = true
					log.Infof("Replicando la request Set a %s.", suc.IP)
					err := node.RPC.Set(suc, req)
					if err != nil {
						log.Errorf("Error replicando la request %s.\n%s", suc.IP, err.Error())
					}
				}()
			}
		}

		log.Info("Se resolvio la request Set de forma exitosa ")
		return &chord.SetResponse{}, nil
	}

	// Si el sucesor es diferente del nodo actual, reenviamos la solicitud al sucesor
	return &chord.SetResponse{}, node.RPC.Set(successor, req)
}

//Ok
func (node *Node) Get(ctx context.Context, req *chord.GetRequest) (*chord.GetResponse, error) {
	log.Infof("Get: key=%s.", req.Ident)
	// Obtenemos el sucesor del nodo actual
	successor, err := node.LocateKey(req.Ident)
	if err != nil {
		log.Error("Error getting key." + err.Error())
		return &chord.GetResponse{}, err
	}

	// Si el sucesor es igual al nodo actual, entonces el valor que se quiere es local,
	// por lo que se recupera del diccionario y se devuelve
	if Equals(successor.ID, node.ID) {

		node.dictLock.Lock()
		value, err := node.dictionary.Get(req.Ident)
		node.dictLock.Unlock()
		if err != nil {
			log.Error("Error getting value." + err.Error())
			return &chord.GetResponse{}, err
		}

		return &chord.GetResponse{ResponseFile: value}, nil
	}

	// Si el sucesor es diferente del nodo actual,no es local. Por lo que se reenvia la solicitud al sucesor
	return node.RPC.Get(successor, req)
}

//Ok
// Elimina un fichero del almacenamiento
func (node *Node) Delete(ctx context.Context, req *chord.DeleteRequest) (*chord.DeleteResponse, error) {
	log.Infof("Eliminar: key=%s.", req.Ident)
	// Obtenemos el sucesor del nodo actual
	successor, err := node.LocateKey(req.Ident)
	if err != nil {
		log.Error("Error getting key." + err.Error())
		return &chord.DeleteResponse{}, err
	}

	// Si el sucesor es igual al nodo actual o es una replicas, entonces tenemos el valor
	if Equals(successor.ID, node.ID) || req.Replication {

		node.dictLock.Lock()
		err := node.dictionary.Delete(req.Ident)
		node.dictLock.Unlock()
		if err != nil {
			log.Error("Error getting value." + err.Error())
			return &chord.DeleteResponse{}, err
		}
		if req.Replication {
			log.Info("Resolviendo la request Delete localmente (replicacion).")
		} else {
			// Se bloquea el sucesosr para leer de el, se desbloquea al terminar
			node.sucLock.RLock()
			suc := node.successors.Beg()
			node.sucLock.RUnlock()
			// Si el sucesor es distinto a este nodo, se replica la request
			if !Equals(suc.ID, node.ID) {
				go func() {
					req.Replication = true
					log.Infof("Replicando la request Delete a %s.", suc.IP)
					err := node.RPC.Delete(suc, req)
					if err != nil {
						log.Errorf("Error replicando la request %s.\n%s", suc.IP, err.Error())
					}
				}()
			}
		}

		log.Info("Se resolvio la request Delete de forma exitosa ")
		return &chord.DeleteResponse{}, nil
	}

	// Si el sucesor es diferente del nodo actual, reenviamos la solicitud al sucesor
	return &chord.DeleteResponse{}, node.RPC.Delete(successor, req)
}

//Ok
// Partition devuelve todos los pares <key, values>  de este almacenamiento local
func (node *Node) Partition(ctx context.Context, req *chord.PartitionRequest) (*chord.PartitionResponse, error) {
	log.Info("Obteniendo todos los archivos y etiquetas en el almacenamiento local.")

	//Bloquea el predecesor para leer de el, lo desbloquea al terminar
	node.predLock.RLock()
	pred := node.predecessor
	node.predLock.RUnlock()

	//Bloquea el predecesor para leer de el, lo desbloquea al terminar
	node.dictLock.RLock()
	// Obtiene los pares <key, value> del almacenamiento.
	in, out, err := node.dictionary.Partition(pred.ID, node.ID)
	node.dictLock.RUnlock()
	if err != nil {
		log.Error("Error obteniendo los archivos del almacenamiento local.")
		return &chord.PartitionResponse{}, errors.New("error obteniendo los archivos del almacenamiento local.\n" + err.Error())
	}
	/*
		Devuelve el diccionario correspondiente al almacenamiento local de este nodo, y el correspondiente
		al almacenamiento local de replicacion
	*/
	return &chord.PartitionResponse{InFiles: in, OutFiles: out}, nil
}

// Ok
// Extend agrega al diccionario local un nuevo conjunto de pares<key, values> .
func (node *Node) Extend(ctx context.Context, req *chord.ExtendRequest) (*chord.ExtendResponse, error) {
	// Si no hay llaves que agregar, regresa.
	if req.ExtendFiles == nil || len(req.ExtendFiles) == 0 {
		return &chord.ExtendResponse{}, nil
	}

	log.Info("Agregando nuevos elementos al almacenamiento local.")

	// Bloquea el diccionario para escribir en el , al terminar se desbloquea
	node.dictLock.Lock()
	// Agrega los nuevos archivos al almacenamiento.
	err := node.dictionary.Extend(req.ExtendFiles)
	// Agrega las nuevas etiquetas al almacenamiento.
	node.dictLock.Unlock()
	if err != nil {
		log.Error("Error agregando los archivos al almacenamiento local.")
		return &chord.ExtendResponse{}, errors.New("error agregando los archivos al almacenamiento local\n" + err.Error())
	}
	return &chord.ExtendResponse{}, nil
}

//OK
// Discard a list of keys from local storage dictionary.
func (node *Node) Discard(ctx context.Context, req *chord.DiscardRequest) (*chord.DiscardResponse, error) {
	log.Info("Descartando llaves desde el diccionario de almacenamiento local.")

	//Bloquea el diccionario para escribir en el, se desbloquea el final
	node.dictLock.Lock()

	// Elimina los archivos del almacenamiento.
	err := node.dictionary.Discard(req.DiscardFiles)
	node.dictLock.Unlock()

	if err != nil {
		log.Error("Error descartando las llaves del diccionario de almacenamiento.")
		return &chord.DiscardResponse{}, errors.New("error descartando las llaves del diccionario de almacenamiento\n" + err.Error())
	}
	return &chord.DiscardResponse{}, nil
}

// Estos dos ultimos metodos se deben revisar, lo mas probable es que sean innecesarios
// Ok
// Elimina la informacion referente a un archivo de una etiqueta
func (node *Node) DeleteElemn(ctx context.Context, req *chord.DeleteElemnRequest) (*chord.DeleteElemnResponse, error) {
	log.Infof("Eliminando de key=%s los elementos %s", req.Ident, req.RemoveFile)
	// Obtenemos el sucesor del nodo actual
	successor, err := node.LocateKey(req.Ident)
	if err != nil {
		log.Error("Error eliminando los valores." + err.Error())
		return &chord.DeleteElemnResponse{}, err
	}

	// Si el sucesor es igual al nodo actual o es una replicas, entonces tenemos el valor
	if Equals(successor.ID, node.ID) || req.Replication {
		err := node.dictionary.DeleteElemn(req.Ident, req.RemoveFile)
		if err != nil {
			log.Error("Error eliminando los valores." + err.Error())
			return &chord.DeleteElemnResponse{}, err
		}
		if req.Replication {
			log.Info("Resolviendo la request DeleteElemn localmente (replicacion).")
		} else {
			// Se bloquea el sucesosr para leer de el, se desbloquea al terminar
			node.sucLock.RLock()
			suc := node.successors.Beg()
			node.sucLock.RUnlock()
			// Si el sucesor es distinto a este nodo, se replica la request
			if !Equals(suc.ID, node.ID) {
				go func() {
					req.Replication = true
					log.Infof("Replicando la request DeleteElemn a %s", suc.IP)
					err := node.RPC.DeleteElemn(suc, req)
					if err != nil {
						log.Errorf("Error replicando la request %s.\n%s", suc.IP, err.Error())
					}
				}()
			}
		}

		log.Info("Se resolvio la request DeleteElemn de forma exitosa ")
		return &chord.DeleteElemnResponse{}, nil
	}

	// Si el sucesor es diferente del nodo actual, reenviamos la solicitud al sucesor
	return &chord.DeleteElemnResponse{}, node.RPC.DeleteElemn(successor, req)

}

//Ok
func (node *Node) SetElemn(ctx context.Context, req *chord.SetElemRequest) error {
	log.Infof("Almacenando en key=%s los archivos=%s", req.Ident, req.SetFile)
	// Obtenemos el sucesor del nodo actual
	successor, err := node.LocateKey(req.Ident)
	if err != nil {
		log.Error("Error getting key." + err.Error())
		return err
	}

	// Si el sucesor es igual al nodo actual o es una replicas, entonces tenemos el valor
	if Equals(successor.ID, node.ID) || req.Replication {
		err := node.dictionary.SetElem(req.Ident, req.SetFile)
		if err != nil {
			log.Error("Error almacenando el valor." + err.Error())
			return err
		}
		if req.Replication {
			log.Info("Resolviendo la request SetElemn localmente (replicacion).")
		} else {
			// Se bloquea el sucesosr para leer de el, se desbloquea al terminar
			node.sucLock.RLock()
			suc := node.successors.Beg()
			node.sucLock.RUnlock()
			// Si el sucesor es distinto a este nodo, se replica la request
			if !Equals(suc.ID, node.ID) {
				go func() {
					req.Replication = true
					log.Info("Replicando la request SetElemn a %s.", suc.IP)
					err := node.RPC.SetElemn(suc, req)
					if err != nil {
						log.Errorf("Error replicando la request %s.\n%s", suc.IP, err.Error())
					}
				}()
			}
		}

		log.Info("Se resolvio la request SetElemn de forma exitosa ")
		return nil
	}

	// Si el sucesor es diferente del nodo actual, reenviamos la solicitud al sucesor
	return node.RPC.SetElemn(successor, req)

}
