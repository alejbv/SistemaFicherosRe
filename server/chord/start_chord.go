package chord

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alejbv/SistemaFicherosRe/chord"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

/*
Se inicializa el servidor como un nodo del chord, se inicializan los servicios de la
capa de transporte y los hilos que periodicamente estabilizan el servidor
*/
func (node *Node) Start() error {
	log.Info("Iniciando el servidor...")

	// Si este nodo servidor está actualemnte en ejecución se reporta un error
	if IsOpen(node.shutdown) {
		log.Error("Error iniciando el servidor: este nodo esta actualemnte en ejecución.")
		return errors.New("error iniciando el servidor: este nodo esta actualemnte en ejecución")
	}

	node.shutdown = make(chan struct{}) // Reporta que el nodo esta corriendo

	ip := GetOutboundIP()                    // Obtiene la IP de este nodo.
	address := ip.String() + ":" + node.Port // Se obtiene la direccion del nodo.
	log.Infof("Dirección del nodo en %s.", address)

	id, err := HashKey(address, node.config.Hash) // Obtiene el ID correspondiente a la direccion.
	if err != nil {
		log.Error("Error iniciando el nodo: no se puede obtener el hash de la dirección del nodo.")
		return errors.New("error iniciando el nodo: no se puede obtener el hash de la dirección del nodo\n" + err.Error())
	}

	node.ID = id          // Se le asigna al nodo el ID correspondiente.
	node.IP = ip.String() //Se le asigna al nodo el IP correspondiente.
	//currentPath := node.dictionary.GetPath()
	//node.dictionary.ChangePath(currentPath + string(node.ID) + "/")

	// Empezando a escuchar en la dirección correspondiente.
	log.Debug("Tratando de escuchar en la dirección correspondiente.")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("Error iniciando el servidor: no se puede escuchar en la dirección %s.", address)
		return fmt.Errorf(
			fmt.Sprintf("error iniciando el servidor: no se puede escuchar en la dirección %s\n%s", address, err.Error()))
	}
	log.Infof("Escuchando en  %s.", address)

	node.successors = NewQueue[chord.Node](node.config.StabilizingNodes) // Se crea la cola de sucesores.
	node.successors.PushBack(node.Node)                                  // Se establece a este nodo como su propio sucesor.
	node.predecessor = node.Node                                         // Se establece este nodo como su propio predecesor.
	node.fingerTable = NewFingerTable(node.config.HashSize)              // Se crea la finger table.
	node.server = grpc.NewServer(node.config.ServerOpts...)              // Se establece  el nodo como un servidor.
	node.sock = listener.(*net.TCPListener)                              // Se almacena el socket.
	if err != nil {
		log.Errorf("Error iniciando el servidor: el diccionario no es valido.")
		return errors.New("error iniciando el servidor: el diccionario no es valido\n" + err.Error())
	}

	chord.RegisterChordServer(node.server, node) // Se registra a este servidor como un servidor chord.
	log.Info("Registrado nodo como servidor chord.")

	err = node.RPC.Start() // Se empiezan los servicios RPC (capa de transporte).
	if err != nil {
		log.Error("Error iniciando el servidor: no se puede arrancar la capa de transporte.")
		return errors.New("error iniciando el servidor: no se puede arrancar la capa de transporte\n" + err.Error())
	}

	discovered, err := node.NetDiscover(ip) // Descubre la red chord, de existir
	if err != nil {
		log.Error("Error iniciando el servidor:  no se pudo descubrir una red para conectarse.")
		return errors.New("error iniciando el servidor:  no se pudo descubrir una red para conectarse\n" + err.Error())
	}

	// Empezando el servicio en el socket abierto
	go node.Listen()

	// Si existe una red de chord existente.
	if discovered != "" {
		err = node.Join(&chord.Node{IP: discovered, Port: node.Port}) // Se une a la red descubierta.
		if err != nil {
			log.Error("Error uniendose a la red de servidores de chord.")
			return errors.New("error uniendose a la red de servidores de chord.\n" + err.Error())
		}
	} else {
		// En otro caso, Se crea.
		log.Info("Creando el anillo del chord.")
	}

	// Se empiezan los hilos de ejecucion periodica.
	go node.PeriodicallyCheckPredecessor()
	go node.PeriodicallyCheckSuccessor()
	go node.PeriodicallyStabilize()
	go node.PeriodicallyFixSuccessor()
	go node.PeriodicallyFixFinger()
	go node.PeriodicallyFixStorage()
	go node.BroadListen()

	log.Info("Servidor Iniciado.")
	return nil
}

/*
Para el nodo servidor al parar el servicio de la capa de transporte (la parte cliente del chord) y reportando
al nodo que los servicios están apagados. Para hacer que los hilos se paren eventualmente.
Entonces, conecta el sucesor de este nodo y su predecesor directamente, antes de dejar el anillo
*/
func (node *Node) Stop() error {
	log.Info("Cerrando el servidor...")

	// Si el nodo servidor actualmente no está en ejecucion, reporta un error.
	if !IsOpen(node.shutdown) {
		log.Error("Error parando el servidor: este nodo servidor ya está caido.")
		return errors.New("error parando el servidor: este nodo servidor ya está caido")
	}

	//Bloquea el sucesor para leer de el, se desbloquea al terminar
	node.sucLock.RLock()
	suc := node.successors.Beg()
	node.sucLock.RUnlock()

	// Bloquea el predecesor para leer de el, se desbloquea al terminar
	node.predLock.RLock()
	pred := node.predecessor
	node.predLock.RUnlock()

	// Si el nodo es distinto a su predecesor y a su sucesor.
	if !Equals(node.ID, suc.ID) && !Equals(node.ID, pred.ID) {
		// Cambia el predecesor del nodo sucesor del actual nodo a su predecesor.
		err := node.RPC.SetPredecessor(suc, pred)
		if err != nil {
			log.Errorf("Error parando el servidor: error estableciendo el nuevo predecesor del sucesor en  %s.", suc.IP)
			return fmt.Errorf(
				fmt.Sprintf("error parando el servidor: error estableciendo el nuevo predecesor del sucesor en  %s\n.%s",
					suc.IP, err.Error()))
		}

		// Cambia el sucesor del predecesor de este nodo a su sucesor.
		err = node.RPC.SetSuccessor(pred, suc)
		if err != nil {
			log.Errorf("Error parando el servidor: error estableciendo el nuevo sucesor del predecesor en %s.", pred.IP)
			return fmt.Errorf(
				fmt.Sprintf("error parando el servidor: error estableciendo el nuevo sucesor del predecesor en %s\n.%s",
					pred.IP, err.Error()))
		}
	}

	err := node.RPC.Stop() // Para los servicios de RPC (capa de transporte)
	if err != nil {
		log.Error("Error parando el servidor: no se puede parar la capa de transporte.")
		return errors.New("error parando el servidor: no se puede parar la capa de transporte\n" + err.Error())
	}

	node.server.Stop() // Para el nodo servidor.

	close(node.shutdown) // Reporta que el nodo servidor está cada
	log.Info("Servidor cerrado.")
	return nil
}

func (node *Node) Listen() {
	log.Info("Empezando el servicio en el socket abierto.")
	err := node.server.Serve(node.sock)
	if err != nil {
		log.Errorf("No se puede dar el servicio en  %s.\n%s", node.IP, err.Error())
		return
	}
}

/*
Join es para unir este nodo a un anillo Chord previamente existente a traves de otro nodo
previamente descubierto en el broadcast. Para unir este nodo al anillo el sucesor inmediato
de este nodo(Por su ID) es buscado
*/
func (node *Node) Join(knownNode *chord.Node) error {
	log.Info("Uniendose al anillo chord .")
	/*
		Si el nodo es nil entonces se devuelve error. Al menos un nodo del anillo debe ser
		conocido
	*/
	if knownNode == nil {
		log.Error("Error uniendose al anillo chord: el nodo conocido no puede ser nulo.")
		return errors.New("error uniendose al anillo chord: el nodo conocido no puede ser nulo")
	}

	log.Infof("La direccion del nodo es: %s.", knownNode.IP)
	// Busca el sucesor inmediato de esta ID en el anillo
	suc, err := node.RPC.FindSuccessor(knownNode, node.ID)
	if err != nil {
		log.Error("Error uniendose al anillo chord: no se puede encontrar el sucesor de esta ID.")
		return errors.New("error uniendose al anillo chord: no se puede encontrar el sucesor de esta ID.\n" + err.Error())
	}
	// Si la ID obtenida es exactamente la ID de este nodo, entonces este nodo ya existe.
	if Equals(suc.ID, node.ID) {
		log.Error("Error uniendose al anillo chord: un nodo con esta ID ya existe.")
		return errors.New("error uniendose al anillo chord: un nodo con esta ID ya existe")
	}

	// Bloquea el sucesor para escribir en el, al terminar se desbloquea.
	node.sucLock.Lock()
	node.successors.PushBeg(suc) // Actualiza el sucesor de este nodo con el obtenido.
	node.sucLock.Unlock()
	log.Infof("Union exitosa. Nodo sucesor en %s.", suc.IP)
	return nil
}

func (node *Node) FindIDSuccessor(id []byte) (*chord.Node, error) {
	log.Trace("Buscando el sucesor de ID .")

	// Si la ID es nula reporta un error.
	if id == nil {
		log.Error("Error buscando el sucesor: ID no puede ser nula.")
		return nil, errors.New("error buscando el sucesor: ID no puede ser nula")
	}

	// Buscando el más cercano, en esta fingertable, con ID menor que esta ID.
	pred := node.ClosestFinger(id)

	// Si el nodo que se esta buscando es exactamente este, se devuelve su sucesor.
	if Equals(pred.ID, node.ID) {
		// Bloquea el sucesor para leer en el, lo desbloque al terminar.
		node.sucLock.RLock()
		suc := node.successors.Beg()
		node.sucLock.RUnlock()
		return suc, nil
	}

	// Si es diferente, busca el sucesor de esta ID
	// del nodo remoto obtenido.
	suc, err := node.RPC.FindSuccessor(pred, id)
	if err != nil {
		log.Errorf("Error buscando el sucesor de la ID en %s.", pred.IP)
		return nil, fmt.Errorf(
			fmt.Sprintf("error buscando el sucesor de la ID en %s.\n%s", pred.IP, err.Error()))
	}
	// Devolver el sucesor obtenido.
	log.Trace("Encontrado el sucesor de la ID.")
	return suc, nil
}

/*
LocateKey localiza el nodo correspondiente a una llave determinada.
Para eso, el obtiene el hash de la llave para obtener el correspondiente ID, entonces busca por
el sucesor más cercano a esta ID en el anillo. dado que este es el nodo al que le corresponde la llave
*/
func (node *Node) LocateKey(key string) (*chord.Node, error) {
	log.Tracef("Localizando la llave: %s.\n", key)

	// Obteniendo el ID relativo a esta llave
	id, err := HashKey(key, node.config.Hash)
	if err != nil {
		log.Errorf("Error generando el hash de la llave: %s.\n", key)
		return nil, fmt.Errorf(fmt.Sprintf("error generando el hash de la llave: %s.\n%s", key, err.Error()))
	}

	// Busca y obtiene el sucesor de esta ID
	suc, err := node.FindIDSuccessor(id)
	if err != nil {
		log.Errorf("Error localizando el sucesor de esta ID: %s.\n", key)
		return nil, fmt.Errorf(fmt.Sprintf("error localizando el sucesor de esta ID: %s.\n%s", key, err.Error()))
	}

	log.Trace("Localizacion de la llave exitosa.")
	return suc, nil
}

// ClosestFinger busca el nodo mas cercano que precede a esta ID.
func (node *Node) ClosestFinger(ID []byte) *chord.Node {
	log.Trace("Buscando el nodo mas cercano que precede a esta ID.\n")
	defer log.Trace("Nodo mas cercano encontrado.\n")

	// Itera sobre la finger table en sentido contrario y regresa el nodo mas cercano
	// tal que su ID este entre la ID del nodo actual y la ID dada.
	for i := len(node.fingerTable) - 1; i >= 0; i-- {
		node.fingerLock.RLock()
		finger := node.fingerTable[i]
		node.fingerLock.RUnlock()

		if finger != nil && Between(finger.ID, node.ID, ID) {
			return finger
		}
	}

	// Si ningun otro cumple las condiciones se devuelve este.
	return node.Node
}

// Posible Metodo a modificar
// AbsorbPredecessorKeys agrega las llaves replicadas del anterior predecesor a este nodo.

//Ok
// Se debe modificar para modificarle a a las llaves la nueva ubicacion del archivo
func (node *Node) AbsorbPredecessorKeys(old *chord.Node) {
	// Si el anterior predecesor no es este nodo.
	if !Equals(old.ID, node.ID) {
		log.Debug("Agregando las llaves del predecesor.")

		// Bloquea el sucesor de este nodo para leer de el, despues se desbloquea.
		node.sucLock.RLock()
		suc := node.successors.Beg()
		node.sucLock.RUnlock()

		// De existir el sucesor, transfiere las llaves del anterior predecesor a el, para mantener la replicacion.
		if !Equals(suc.ID, node.ID) {
			// Bloquea el diccionario para leer de el, al terminar se desbloquea.
			node.dictLock.RLock()
			// Obtiene la informacion de las etiquetas  del predecesor.
			in, out, err := node.dictionary.Partition(old.ID, node.ID)
			node.dictLock.RUnlock()
			if err != nil {
				log.Errorf("Error obteniendo los archivos del viejo predecesor.\n%s", err.Error())
				return
			}

			log.Debug("Transferiendo las llaves del predecesor al sucesor.")
			log.Debugf("Llaves a transferir: %s", Keys(out))
			log.Debugf("LLaves restantes: %s", Keys(in))

			// Trasfiere las del predecesor al nodo sucesor.
			err = node.RPC.Extend(suc, &chord.ExtendRequest{ExtendFiles: out})
			if err != nil {
				log.Errorf("Error transfiriendo las llaves al sucesor.\n%s", err.Error())
				return
			}
			log.Debug("Transferencia de las llaves al sucesor exitosa.")

		}
	}
}

// Ok
// DeletePredecessor elimina los archivos replicadoss del viejo sucesor en este nodo.
// Devuelve los archivos actuales, los replicados actuales y las eliminados.
func (node *Node) DeletePredecessorKeys(old *chord.Node) (map[string][]byte, map[string][]byte, map[string][]byte, error) {
	//Bloquea el predecesor para leer de el, se desbloquea el terminar
	node.predLock.Lock()
	pred := node.predecessor
	node.predLock.Unlock()

	//Bloquea el diccionario para leer y escribir en el, se desbloquea el terminar la funcion
	node.dictLock.Lock()
	defer node.dictLock.Unlock()

	// Obtiene los archivos
	in, out, err := node.dictionary.Partition(pred.ID, node.ID)
	if err != nil {
		log.Error("Error obteniendo los archivos correspondientes al nuevo predecesor.")
		return nil, nil, nil, errors.New("error obteniendo los archivos correspondientes al nuevo predecesor\n" + err.Error())
	}
	// Si el antiguo predecesor es distinto al nodo actual, elimina los archivos del viejo predecesor en este nodo.
	if !Equals(old.ID, node.ID) {
		// Obtiene los archivos a eliminar
		_, deleted, err := node.dictionary.Partition(old.ID, node.ID)
		if err != nil {
			log.Error("Error obteniendo los archivos replicados del antiguo predecesor en este nodo.")
			return nil, nil, nil, errors.New(
				"error obteniendo los archivos replicado del antiguo predecesor en este nodo\n" + err.Error())
		}

		// Elimina los archivos del anterior predecesor
		err = node.dictionary.Discard(Keys(out))
		if err != nil {
			log.Error("Error eliminando las llaves del anterior predecesor en este nodo.")
			return nil, nil, nil, errors.New("error eliminando las llaves del anterior predecesor en este nodo\n" + err.Error())
		}

		log.Debug("Eliminacion exitosa de las llaves replicadas del anterior predecesor en este nodo.")
		return in, out, deleted, nil
	}
	return in, out, nil, nil
}

// Ok
// UpdatePredecessorKeys actualiza el nuevo predecesor con la clave correspondiente.
func (node *Node) UpdatePredecessorKeys(old *chord.Node) {
	//  Bloquea el predecesor para leer de el, lo desbloquea al terminar.
	node.predLock.Lock()
	pred := node.predecessor
	node.predLock.Unlock()

	//  Bloquea el sucesor para escribir en el, lo desbloquea al terminar.
	node.sucLock.Lock()
	suc := node.successors.Beg()
	node.sucLock.Unlock()

	// Si el nuevo predecesor no es el actual nodo.
	if !Equals(pred.ID, node.ID) {
		// Transfiere las claves correspondientes al nuevo predecesor.
		in, out, deleted, err := node.DeletePredecessorKeys(old)

		if err != nil {
			log.Errorf("Error actualizando los archivos del nuevo predecesor.\n%s", err.Error())
			return
		}

		log.Debug("Transfiriendo los ficheros del antiguo predecesor al sucesor.")
		log.Debugf("LLaves a transferir: %s", Keys(out))
		log.Debugf("LLaves restantes: %s", Keys(in))
		log.Debugf("LLaves a eliminar: %s", Keys(deleted))

		// Construye el diccionario del nuevo predecesor, al transferir sus claves correspondientes.
		err = node.RPC.Extend(pred, &chord.ExtendRequest{ExtendFiles: out})
		if err != nil {
			log.Errorf("Error transfiriendo la informacion al nuevo predecesor.\n%s", err.Error())
			/*
				De existir archivos eliminados, se reinsertan en el almacenamiento de este nodo
				para prevenir perdida de informacion
			*/
			if deleted != nil {
				//  Bloquea el diccionario para escribir en el, lo desbloquea al terminar.
				node.dictLock.Lock()
				err = node.dictionary.Extend(deleted)
				node.dictLock.Unlock()
				if err != nil {
					log.Errorf("Error reinsertando etiquetas eliminadas al diccionario.\n%s", err.Error())
				}
			}
			return
		}
		log.Debug("Transferencia exitosa de las llaves al nuevo predecesor.")
		/*
			De existir el sucesor, y es diferente del nuevo predecesor, se eliminan las llaves transferidas
			del almacenamiento del sucesor
		*/
		if !Equals(suc.ID, node.ID) && !Equals(suc.ID, pred.ID) {
			err = node.RPC.Discard(suc, &chord.DiscardRequest{DiscardFiles: Keys(out)})
			if err != nil {
				log.Errorf("Error eliminando las claves replicadas en el nodo sucesor en %s.\n%s", suc.IP, err.Error())
				/*
					De haber informacion eliminada, se reinserta para prevenir la perdida de informacion
				*/
				if deleted != nil {
					// Bloquea el diccionario para leer y escribir en el, al terminar se desbloquea.
					node.dictLock.Lock()
					err = node.dictionary.Extend(deleted)
					node.dictLock.Unlock()
					if err != nil {
						log.Errorf("Error reinsertando las etiquetas en este nodo.\n%s", err.Error())
					}
				}
				return
			}
			log.Debug("Eliminacion exitosa de las llaves replicadas del antiguo predecesor en el nodo sucesor.")
		}
	}
}

// Ok
// UpdateSuccessorKeys actualiza el nuevo sucesor replicando la informacion del nodo actual.
func (node *Node) UpdateSuccessorKeys() {
	//Bloquea el predecesor para leer de el, al terminar se desbloquea.
	node.predLock.RLock()
	pred := node.predecessor
	node.predLock.RUnlock()

	//Bloquea el sucesor para escribir en el, al terminar se desbloquea.
	node.sucLock.Lock()
	suc := node.successors.Beg()
	node.sucLock.Unlock()

	// Si el sucesor es distinto de este nodo
	if !Equals(suc.ID, node.ID) {
		//Bloquea el diccionario para leer de el, al terminar se desbloquea.
		node.dictLock.RLock()
		in, out, err := node.dictionary.Partition(pred.ID, node.ID) // Obtiene los archivos del nodo
		node.dictLock.RUnlock()
		if err != nil {
			log.Errorf("Error obtaining this node keys.\n%s", err.Error())
			return
		}

		log.Debug("Transferring this node keys to the successor.")
		log.Debugf("Keys to transfer: %s", Keys(in))
		log.Debugf("Remaining keys: %s", Keys(out))

		// Transfer this node keys to the new successor, to update it.
		err = node.RPC.Extend(suc, &chord.ExtendRequest{ExtendFiles: in})
		if err != nil {
			log.Errorf("Error transferring keys to the successor at %s.\n%s", suc.IP, err.Error())
			return
		}

		log.Trace("Transferencia de llaves al sucesor exitosa.")
	}
}

// BroadListen espera por mensajes de broadcast.
func (node *Node) BroadListen() {

	// Resolver para poder escuchar la respuesta al broadcast
	listAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8888")
	if err != nil {
		log.Error("Error al vincular el puerto 8888 para mensajes entrantes en la parte del servidor")
		return
	}

	// Tratando de escuchar al puerto en uso.
	listen, err := net.ListenUDP("udp", listAddr)
	if err != nil {
		log.Errorf("Error al crear el objeto que recibe las transmisiones: %s\n", err.Error())
		return
	}
	defer listen.Close()
	log.Info("Se logro crear exitosamente el listener asociado al puerto 8888")

	// Empieza a escuchar mensajes
	log.Info("El nodo empezo a escuchar")
	for {
		// Si el nodo servidor está caido, regresa
		if !IsOpen(node.shutdown) {
			log.Info("El nodo dejo de escuchar")
			return
		}

		// Crea el buffer para almacenar el mensaje
		buf := make([]byte, 15000)
		// Espera por el mensaje
		n, address, err := listen.ReadFromUDP(buf[0:])

		// Se cambia el puerto al que se va a mandar la informacion
		address.Port = 8838

		if err != nil {
			log.Errorf("Error de mensaje de broadcast entrante.\n%s", err.Error())
			continue
		}
		if n != 0 {
			log.Infof("Llego informacion :%s\n", string(buf))
		}

		log.Infof("Respuesta al mensaje entrante. %s enviar esto: %s", address, buf[:n])

		// Si el mensaje entrante es el especifico, se le responde con la respuesta en especifico.
		if string(buf[:n]) == "Chord?" {
			log.Info("Llego un mensaje de un nuevo nodo")
			_, err = listen.WriteToUDP([]byte("Yo soy chord"), address)
			if err != nil {
				log.Errorf("Error respondiendo al mensaje de broadcast.\n%s", err.Error())
				continue
			}
		}
	}
}

// NetDiscover trata de encontrar si existe en la red algun otro anillo chord.
func (node *Node) NetDiscover(ip net.IP) (string, error) {
	// La dirección de broadcast.
	ip[2] = 255
	ip[3] = 255
	broadcast := ip.String() + ":8888"

	//Resuelve la dirección a la que se va a hacer broadcast.
	outAddr, err := net.ResolveUDPAddr("udp", broadcast)
	if err != nil {
		log.Errorf("Error resolviendo la direccion de broadcast %s.", broadcast)
		return "", err
	}

	// Tratando de escuchar al puerto en uso.
	conn, err := net.ListenUDP("udp", outAddr)
	if err != nil {
		log.Errorf("Error al escuchar a la dirección %s.", broadcast)
		return "", err
	}
	defer conn.Close()

	log.Infof("Resuelta direccion UPD broadcast:%s.", broadcast)

	//Se crea un objeto para recibir la respuesta del otro servidor
	listAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:8838")
	if err != nil {
		log.Error("Error resolviendo la direccion para escuchar la respuesta 0.0.0.0:8838.\n")
		return "", err
	}

	// Build listining connections
	listen, err := net.ListenUDP("udp", listAddr)
	if err != nil {
		log.Errorf("Error al escuchar a la direccion de respuesta ")
		return "", err
	}
	defer listen.Close()

	// Enviando el mensaje
	message := []byte("Chord?")
	_, err = conn.WriteToUDP(message, outAddr)
	if err != nil {
		log.Errorf("Error enviando el mensaje de broadcast a la dirección%s.", broadcast)
		return "", err
	}

	log.Info("Terminado el mensaje broadcast.")
	top := time.Now().Add(10 * time.Second)

	log.Info("Esperando por respuesta.")

	for top.After(time.Now()) {
		// Creando el buffer para almacenar el mensaje.
		buff := make([]byte, 15000)

		// Estableciendo el tiempo de espera para mensajes entrantes.
		err = listen.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			log.Error("Error establecientdo el tiempo de espera para mensajes entrantes.")
			return "", err
		}

		// Esperando por un mensaje.
		n, address, err := listen.ReadFromUDP(buff)
		if err != nil {
			log.Errorf("Error leyendo el mensaje entrante.\n%s", err.Error())
			continue
		}

		log.Infof("Mensaje de respuesta entrante. %s enviado a: %s", address, buff[:n])

		if string(buff[:n]) == "Yo soy chord" {
			return strings.Split(address.String(), ":")[0], nil
		}
	}

	log.Info("Tiempo de espera por respuesta pasado.")
	return "", nil
}
