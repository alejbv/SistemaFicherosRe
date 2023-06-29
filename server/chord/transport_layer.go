package chord

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alejbv/SistemaFicherosRe/chord"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// RemoteServices permite a un nodo a interactuar con otros nodos en el anillo, como un cliente a sus servicios.
type RemoteServices interface {
	Start() error // Empieza el servicio.
	Stop() error  // Para el servicio.

	// GetPredecessor devuelve el nodo que se cree que es el actual predecesor de un nodo remoto.
	GetPredecessor(*chord.Node) (*chord.Node, error)
	// GetSuccessor r devuelve el nodo que se cree que es el actual sucesor de un nodo remoto.
	GetSuccessor(*chord.Node) (*chord.Node, error)
	// SetPredecessor establece el predecesor de un nodo remoto.
	SetPredecessor(*chord.Node, *chord.Node) error
	// SetSuccessor sestablece el sucesor de un nodo remoto.
	SetSuccessor(*chord.Node, *chord.Node) error
	// FindSuccessor encuentra el nodo que sucede a la ID, partiendo desde un nodo remoto.
	FindSuccessor(*chord.Node, []byte) (*chord.Node, error)
	// Notify notifica a un nodo remoto de que es posible tenga un nuevo predecesor.
	Notify(*chord.Node, *chord.Node) error
	// Check Comprueba si un nodo remoto esta vivo.
	Check(*chord.Node) error
	// Metodos de la aplicacion
	// Añade un archivo al almacenamiento local del nodo correspondiente
	Set(node *chord.Node, req *chord.SetRequest) error
	// Elimina un fichero del almacenamiento
	Delete(node *chord.Node, req *chord.DeleteRequest) error

	Get(node *chord.Node, req *chord.GetRequest) (*chord.GetResponse, error)
	// Elimina la informacion referente a un archivo de una etiqueta
	DeleteElemn(node *chord.Node, req *chord.DeleteElemnRequest) error

	SetElemn(node *chord.Node, req *chord.SetElemRequest) error
	// Partition  devuelve todos los pares <key, values> en un intervalo dado, del almacenamiento de un nodo remoto
	Partition(node *chord.Node, req *chord.PartitionRequest) (*chord.PartitionResponse, error)

	Extend(node *chord.Node, req *chord.ExtendRequest) error
	// Partition return all <key, values> pairs in a given interval from the storage of a remote node.
	Discard(node *chord.Node, req *chord.DiscardRequest) error
}
type GRPCServices struct {
	*Configuration // Configuraciones de un nodo remoto.

	connections    map[string]*RemoteNode // Diccionario de <address, open connection>.
	connectionsMtx sync.RWMutex           // Bloquea el diccionario para lectura o escritura.

	shutdown chan struct{} // Determina si el servicio esta actualmente en ejecucion
}

// NewGRPCServices crea un nuevo objeto tipo GRPCServices .
func NewGRPCServices(config *Configuration) *GRPCServices {
	// Crea el GRPCServices.
	services := &GRPCServices{
		Configuration: config,
		connections:   nil,
		shutdown:      nil,
	}

	// Devuelve el GRPCServices.
	return services
}

// Empieza el servicio.
func (services *GRPCServices) Start() error {
	log.Info("Empezando el servicio de la capa de transporte...\n")

	// Si los servicios de la capa de tranporte estan en funcionamiento reporta un error .
	if IsOpen(services.shutdown) {
		message := "Error empezando el servicio:los servicios de la capa de tranporte estan en funcionamiento."
		log.Error(message)
		return errors.New(fmt.Sprintln(message))
	}

	services.shutdown = make(chan struct{}) // Reportar que los servicios estan en ejecucion.

	services.connections = make(map[string]*RemoteNode) // Crear el diccionario de <address, open connection>.
	// Empezar hilos periodicos.
	go services.CloseOldConnections() // Comprueba y cierra viejas conexiones.

	log.Info("Servicios de la capa de tranporte en funcionamiento\n")
	return nil
}

/*
Cierra el servicio, reportando que los servicios de la capa de transporte estan apagados
*/
func (services *GRPCServices) Stop() error {
	log.Info(fmt.Sprintln("Parando los servicios de la capa de transporte..."))

	// Si los servicios de la capa de transporte no estan en funcionamiento reporta error.
	if !IsOpen(services.shutdown) {
		message := "Error parando los servicios: estos ya estan cerrados."
		log.Error(message)
		return errors.New(fmt.Sprintln(message))
	}

	close(services.shutdown) // Reportar que los servicios se pararon.
	log.Info("Servicios de la capa de transporte parados.\n")
	return nil
}

// Connect establece una conexion con una direccion remota.
func (services *GRPCServices) Connect(addr string) (*RemoteNode, error) {
	log.Trace("Conectando con la direccion " + addr + ".\n")

	// Comprueba si el servicio esta apagado, en caso de que si se termina y se reporta.
	if !IsOpen(services.shutdown) {
		message := "Error creando la conexion: se deben empezar los servicios de la capa de transporte primero."
		log.Error(message)
		return nil, errors.New(fmt.Sprintln(message))
	}

	services.connectionsMtx.RLock()
	if services.connections == nil {
		services.connectionsMtx.Unlock()
		message := "Error creando la conexion: la tabla de conexion esta vacia."
		log.Error(message)
		return nil, errors.New(fmt.Sprintln(message))
	}
	remoteNode, ok := services.connections[addr]
	services.connectionsMtx.RUnlock()

	// Comprueba si existe una conexion que este en funcionamiento, de ser asi se devuelve dicha conexion .
	if ok {
		log.Trace("Conexion exitosa.\n")
		return remoteNode, nil
	}

	conn, err := grpc.Dial(addr, services.DialOpts...) // De otra forma se crea.
	if err != nil {
		message := "Error creando la conexion.\n"
		log.Error(message)
		return nil, errors.New(message + err.Error())
	}

	client := chord.NewChordClient(conn) // Crear el client chord asociado con la conexion.
	// Se construye el correspondiente nodo remoto.

	remoteNode = &RemoteNode{client,
		addr,
		conn,
		time.Now()}

	services.connectionsMtx.Lock()
	services.connections[addr] = remoteNode
	services.connectionsMtx.Unlock()

	log.Trace(fmt.Sprintln("Conexion exitosa."))
	return remoteNode, nil
}

// CloseOldConnections cierra las viejas conexiones abiertas.
func (services *GRPCServices) CloseOldConnections() {
	log.Trace(fmt.Sprintln("Cerrando viejas conexiones."))

	// Si el servicio esta apagado , cierra todas las conexiones  y regresa.
	if !IsOpen(services.shutdown) {
		services.connectionsMtx.Lock() // Bloquea el diccionario para que no se escriba en el,Se desbloquea despues.
		// Para  cerrar las conexiones con los nodos en el diccionario.
		for _, remoteNode := range services.connections {
			remoteNode.CloseConnection()
		}
		services.connections = nil // Borra el diccionario de conexiones.
		services.connectionsMtx.Unlock()
		return
	}

	services.connectionsMtx.RLock() // Bloquea el diccionario para lectura en el,Se desbloquea despues.
	if services.connections == nil {
		services.connectionsMtx.Unlock()
		log.Error("Error cerrando conexiones: la tabla de conexiones esta vacia.\n")
		return
	}
	services.connectionsMtx.RUnlock()

	services.connectionsMtx.Lock()

	for addr, remoteNode := range services.connections {
		if time.Since(remoteNode.lastActive) > services.MaxIdle {
			remoteNode.CloseConnection()
			delete(services.connections, addr) // Elimina los pares <address, open connection> en el diccionario.
		}
	}
	services.connectionsMtx.Unlock()
	log.Trace(fmt.Sprintln("Viejas conexiones cerradas."))
}

// PeriodicallyCloseConnections periodicamente cierra las viejas conexiones.
func (services *GRPCServices) PeriodicallyCloseConnections() {
	ticker := time.NewTicker(60 * time.Second) // Establece el tiempo de reactivacion de las rutinas.
	for {
		select {
		case <-ticker.C:
			services.CloseOldConnections() // Si se cumple el tiempo, cierra todas las conexiones viejas.
		case <-services.shutdown:
			services.CloseOldConnections() // Si er servicio está caído, cierra todas las conexions y cierra el hilo.
			return
		}
	}
}

// Set almacena  un fichero en el almacenamiento local .
func (services *GRPCServices) Set(node *chord.Node, req *chord.SetRequest) error {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return err
	}
	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.Set(ctx, req)
	return err
}

func (services *GRPCServices) Get(node *chord.Node, req *chord.GetRequest) (*chord.GetResponse, error) {
	if node == nil {
		return nil, errors.New(fmt.Sprintln("No se puede establecer conexion con un nodo vacio."))
	}
	// Estableciendo conexion con un nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return nil, err
	}
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()
	// Se devuelve el resultado de la llamada remota
	return remoteNode.Get(ctx, req)
}

func (services *GRPCServices) Delete(node *chord.Node, req *chord.DeleteRequest) error {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return err
	}
	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.Delete(ctx, req)
	return err
}

func (services *GRPCServices) SetElemn(node *chord.Node, req *chord.SetElemRequest) error {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return err
	}
	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.SetElem(ctx, req)
	return err
}

func (services *GRPCServices) DeleteElemn(node *chord.Node, req *chord.DeleteElemnRequest) error {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return err
	}
	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.DeleteElemn(ctx, req)
	return err
}

// GetPredecessor devuelve el nodo que se cree que es el actual predecesor de un nodo remoto
func (services *GRPCServices) GetPredecessor(node *chord.Node) (*chord.Node, error) {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return nil, err
	} //Se devuelve el resultado de la llamada remota.

	res, err := remoteNode.GetPredecessor(ctx, &chord.GetPredecessorRequest{})
	return res.GetPredecessor(), err
}

// GetSuccessor returns the node believed to be the current successor of a remote node.
func (services *GRPCServices) GetSuccessor(node *chord.Node) (*chord.Node, error) {
	if node == nil {
		return nil, errors.New(fmt.Sprintln("No se puede establecer una conexion con un nodo vacio."))
	}

	// Estableciendo conexion con un nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return nil, err
	}

	// Se obtiene el contexto de la conexion y el tiempo de espera de esta
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()

	// Devuelve el resultado de la llamada remota.
	res, err := remoteNode.GetSuccessor(ctx, &chord.GetSuccessorRequest{})
	return res.GetSuccessor(), err
}

// SetPredecessor establece el predecesor de un nodo remoto.
func (services *GRPCServices) SetPredecessor(node, pred *chord.Node) error {
	if node == nil {
		return errors.New(fmt.Sprintln("No se puede establecer conexion con un nodo vacio."))
	}

	// Establecida conexion con un nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return err
	}

	//Se obtiene el contexto de una conexion y el tiempo de espera de esta
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()

	// Devuelve el resultado de la llamada remota.
	req := &chord.SetPredecessorRequest{Predecessor: pred}
	_, err = remoteNode.SetPredecessor(ctx, req)
	return err
}

// SetSuccessor establece el sucesor de un nodo remoto.
func (services *GRPCServices) SetSuccessor(node, suc *chord.Node) error {
	if node == nil {
		return errors.New(fmt.Sprintln("No se puede establecer conexion con un nodo vacio."))
	}

	// Establecida conexion con un nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return err
	}

	//Se obtiene el contexto de una conexion y el tiempo de espera de esta
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()

	// Devuelve el resultado de la llamada remota.
	req := &chord.SetSuccessorRequest{Successor: suc}
	_, err = remoteNode.SetSuccessor(ctx, req)
	return err
}

// FindSuccessor encuentra el nodo que sucede a esta ID, empezando desde el nodo remoto.
func (services *GRPCServices) FindSuccessor(node *chord.Node, id []byte) (*chord.Node, error) {
	if node == nil {
		return nil, errors.New(fmt.Sprintln("No se puede establecer una conexion con un nodo nulo."))
	}

	remoteNode, err := services.Connect(node.IP + ":" + node.Port) // Establece la conexion con el nodo remoto.
	if err != nil {
		return nil, err
	}

	// Obtiene el contexto de la conexion y establece  el tiempo de espera de la peticion.
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()

	// Devuelve el resultado de la llamada remota.
	req := chord.FindSuccesorRequest{ID: id}
	res, err := remoteNode.FindSuccessor(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res.GetSuccesor(), nil
}

// Notify notifica a un nodo remoto que es posible que tenga un nuevo predecesor.
func (services *GRPCServices) Notify(node, pred *chord.Node) error {
	if node == nil {
		return errors.New(fmt.Sprintln("No se puede establecer conexion con un nodo vacio."))
	}

	// Estableciendo conexion con un nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return err
	}

	// Obteniendo el contexto de la conexion y el tiempo de espera de la.
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()

	// Devuelve el resultado de la llamada remota.
	req := &chord.NotifyRequest{Notify: pred}
	_, err = remoteNode.Notify(ctx, req)
	return err
}

// Comprueba si un nodo remoto esta vivo.
func (services *GRPCServices) Check(node *chord.Node) error {
	if node == nil {
		return errors.New(fmt.Sprintln("No se puede establecer una conexion con un nodo nulo."))
	}

	// Establce conexion con el nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return err
	}

	// Obtiene el contexto de la conexion y establece el tiempo de duracion de esta.
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	defer cancel()

	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.Check(ctx, &chord.CheckRequest{})
	return err
}

// Partition devuelve todos la informacion  en un intervalo dado del almacenamiento de un nodo remoto.
func (services *GRPCServices) Partition(node *chord.Node, req *chord.PartitionRequest) (*chord.PartitionResponse, error) {
	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return nil, err
	}

	// Devuelve el resultado de la llamada remota.
	return remoteNode.Partition(ctx, req)
}

// Extend manda una lista de pares <llave,valor> al almacenamiento de un nodo remoto
func (services *GRPCServices) Extend(node *chord.Node, req *chord.ExtendRequest) error {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return err
	}
	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.Extend(ctx, req)
	return err
}

func (services *GRPCServices) Discard(node *chord.Node, req *chord.DiscardRequest) error {
	ctx, cancel, remoteNode, err := services.GeneralConnection(node)
	defer cancel()

	if err != nil {
		log.Error("Error iniciando una conexion al nodo remoto")
		return err
	}
	// Devuelve el resultado de la llamada remota.
	_, err = remoteNode.Discard(ctx, req)
	return err
}

func (services *GRPCServices) GeneralConnection(node *chord.Node) (context.Context, context.CancelFunc, *RemoteNode, error) {
	if node == nil {
		return nil, nil, nil, errors.New(fmt.Sprintln("No se puede establecer una conexion con un nodo vacio."))
	}

	// Estableciendo conexion con un nodo remoto
	remoteNode, err := services.Connect(node.IP + ":" + node.Port)
	if err != nil {
		return nil, nil, nil, err
	}

	// Se obtiene el contexto de la conexion y el tiempo de espera de la request.
	ctx, cancel := context.WithTimeout(context.Background(), services.Timeout)
	return ctx, cancel, remoteNode, nil
}
