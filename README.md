# Introducción

Un sistema de ficheros basado en etiquetas es un tipo de sistema de ficheros que permite etiquetar o clasificar archivos y directorios de forma más flexible y dinámica que un sistema de ficheros tradicional basado en árbol de directorios.

En un sistema de ficheros basado en etiquetas, cada archivo o directorio puede tener una o varias etiquetas asociadas, que pueden ser palabras clave, categorías, etiquetas de estado, etiquetas de proyecto, etiquetas de contexto, o cualquier otra información que permita clasificar y encontrar los archivos de forma más eficiente. Las etiquetas pueden ser asignadas o modificadas de forma dinámica por el usuario, sin necesidad de mover o copiar los archivos, lo que permite una mayor flexibilidad y organización.

En ellos ademas es mucho mas facil la busqueda y recuperacion de los archivos ya que solo es necesario saber algunas de las etiquetas asociadas a ese archivo en lugar de la ubicacion exacta de este asi como tambien permite una organización más personalizada y adaptable

En este trabajo se desarrollo un sistema de ficheros basado en etiquetas usando el protocolo chord para usando una arquitectura peer-to-peer (P2P) donde cada nodo en la red actua a su ves tanto como cliente y como servidor. Está diseñado para ser escalable y tolerante a fallos permitiendole manejar un mayor trafico sin perder en eficiencia.
# Instalación
Este proyecto está completamente desarrollado en go. Si bien no es necesario tener este lenguaje instalados
si necesario tener instalado Docker y una imagen del Go, sobre la que se construira una imagen de este proyecto para su posterior uso. Ademas este proyecto usa los paquetes:

- github.com/spf13/cobra
- github.com/spf13/viper
- github.com/golang/protobuf/proto
- google.golang.org/grpc
- github.com/sirupsen/logrus

# Protocolo Chord

El protocolo Chord es un protocolo de comunicación y un algoritmo de enrutamiento descentralizado para la implementación de tablas de hash distribuidas. Es un sistema sencillo, escalable  y está diseñado para funcionar en redes descentralizadas peer-to-peer (P2P). Su objetivo es permitir que cualquier nodo dentro de la red pueda buscar y acceder a un objeto almacenado en cualquier otro nodo, incluso si no se tiene información sobre la ubicación física del objeto o el nodo que lo almacena.

El protocolo Chord es un protocolo de tipo anillo, en el que cada nodo almacena información sobre los nodos que están a su izquierda y a su derecha en el anillo. Cada nodo recibe una identificación única, que se utiliza para asignar el nodo a una posición en el anillo. Las identificaciones se asignan utilizando una función hash, lo que permite que los nodos se distribuyan uniformemente en el anillo.

Cuando un nodo necesita buscar un objeto, utiliza la función hash para determinar la posición del objeto en el anillo y luego utiliza el algoritmo de enrutamiento de Chord para encontrar el nodo que almacena el objeto. El algoritmo de enrutamiento utiliza una tabla de enrutamiento para determinar el siguiente salto en el anillo que se acerca más al nodo que almacena el objeto.

El protocolo Chord es escalable y tolerante a fallos, ya que los nodos pueden unirse y abandonar la red sin afectar significativamente el rendimiento o la disponibilidad del sistema. Además, el protocolo Chord se puede utilizar para implementar una variedad de aplicaciones P2P, como compartir archivos, transmisión de multimedia, y redes sociales descentralizadas, entre otros.

# Tolerancia a Fallos

Para garantizar la tolerancia a fallos en Chord, se utilizan dos técnicas principales:

1. Replicación de datos: Chord replica los datos en varios nodos de la red para garantizar que si un nodo falla, los datos aún estarán disponibles en otros nodos. La cantidad de replicas puede ser configurada y dependerá del grado de tolerancia a fallos que se desee.
2. Redundancia de nodos: Chord utiliza un sistema de anillos virtuales donde cada nodo es responsable de un rango de claves. Si un nodo falla, sus responsabilidades son transferidas a otro nodo en el anillo, garantizando que la red siga funcionando sin interrupciones. Además, Chord utiliza una técnica de "sucesor de respaldo" para tener un nodo de respaldo para cada nodo en la red. Si un nodo falla, el sucesor de respaldo se convierte en el sucesor principal y se garantiza que siempre haya un nodo disponible para manejar las solicitudes de la red

Además de las técnicas de replicación de datos y redundancia de nodos mencionadas anteriormente, el protocolo Chord también utiliza otras estrategias para garantizar la tolerancia a fallos.

Una de ellas es la técnica de "estabilización", que se utiliza para mantener la coherencia de la información de los nodos y actualizar la información de los sucesores y predecesores después de que se produzcan cambios en la red, como la unión o la salida de un nodo. Cuando un nodo cae, inmediatamente los nodos vecinos a el se reajustan de la siguiente forma

1. Cuando un nodo falla, los nodos vecinos en el anillo detectan su ausencia mediante el envío periódico de mensajes de "ping" para verificar si el nodo sigue activo. Si el nodo no responde a los mensajes de ping después de varios intentos, se considera que ha fallado
2. Una vez que se ha detectado la falla del nodo, se selecciona automáticamente el sucesor de respaldo del nodo fallido para tomar su lugar en el anillo. El sucesor de respaldo es el nodo que sigue al nodo fallido en el anillo y se utiliza para garantizar que siempre haya un nodo disponible para manejar las solicitudes de la red.
3. Después de que se ha seleccionado el sucesor de respaldo, los nodos vecinos actualizan su tabla de enrutamiento para dirigir las solicitudes que iban dirigidas al nodo fallido hacia el sucesor de respaldo. De esta manera, las solicitudes de la red pueden continuar siendo atendidas sin interrupciones.
4. Si el nodo fallido tenía réplicas de datos en otros nodos de la red, estas réplicas se actualizan automáticamente para garantizar que los datos sigan estando disponibles en la red.

La técnica de estabilización también ayuda a detectar y corregir cualquier inconsistencia en la información de la red.

Otra técnica utilizada en Chord es la "recuperación de fallas", que se utiliza para recuperar los datos y la funcionalidad de los nodos que han fallado. Cuando un nodo falla, los nodos vecinos en la red detectan su ausencia y toman medidas para recuperar los datos y las responsabilidades del nodo fallido.

# Escalabilidad

El protocolo Chord está diseñado para ser escalable, lo que significa que puede manejar redes distribuidas de gran tamaño y aumentar su capacidad a medida que se agregan nuevos nodos. La escalabilidad en Chord se logra principalmente a través de dos técnicas:

1. Asignación de claves según el hash: Chord utiliza una función de hash para asignar claves a los nodos en el anillo virtual. Esto significa que la ubicación de los datos y los recursos en la red se determina de forma aleatoria, en función de su clave hash. Como resultado, los nodos no necesitan conocer la topología completa de la red para enrutar las solicitudes hacia su destino. En cambio, solo necesitan conocer la ubicación del nodo responsable en el anillo virtual.
2. Enrutamiento basado en saltos de longitud logarítmica: Chord utiliza un esquema de enrutamiento basado en saltos de longitud logarítmica, lo que significa que el número de saltos requeridos para enrutar una solicitud a su destino aumenta de forma logarítmica con el tamaño de la red. Esto significa que la cantidad de mensajes necesarios para enrutar una solicitud no aumenta proporcionalmente con el número de nodos en la red. En cambio, se mantiene relativamente constante, lo que permite que Chord maneje redes de gran tamaño de manera eficiente.

Además de estas técnicas, Chord también utiliza una técnica de "aceleración de arranque" para permitir que los nuevos nodos se unan a la red de manera rápida y eficiente. Cuando un nuevo nodo se une a la red, solicita información sobre el estado actual de la red a un nodo existente y utiliza esta información para construir rápidamente su tabla de enrutamiento.

# Descripcion General
En este sistema cada nodo actua tanto como cliente asi como servidor. Se utiliza gRPC y TCP para asegurar
la comunicacion entre todas las partes involucradas. Para interactuar con el sistema se utiliza la linea
de comandos, y existe dos formas basicas de interactuar con este, se inicializa como servidor con el comando serverStart, una vez echo eso el nodo empieza a buscar en la red local por otro nodo que pertenezca
al sistema, una vez encontrado empieza el proceso para unirse al anillo chord. De no encontrar ninguno se
inicializa el como un anillo con un solo elemento. Los nodos interactuan entre ellos usando la capa de transporte para asegurar la estabilidad o buscar recursos que no esten locales.

En cambio el cliente puede interactuar con el sistema agregando archivos, eliminandolos y recuperando informacion de los archivos usando la interfaz de la linea de comando. Tambien se puede agregar etiquetas o eliminar segun la eleccion del usuario.

En el anillo se almacenan tanto los archivos como las etiquetas. Al subir un archivo se computa su identificador usando sus metadatos(nombre,extension,tamaño) su informacion y la fecha de subida y se almacena en el nodo adecuado todos sus metadatos y la informacion del archivo junto con las etiquetas con
las que se subio. Las etiquetas en cambio se almancena usando solo *Tags.laetiqueta* y se almacena de la misma forma que los archivos solo que la informacion que almacena es la lista de todos los archivos que tienen a esa etiqueta. Cada vez que un archivo se sube se replica su informacion a su nodo sucesor.
# CLI
Usando los github.com/spf13/cobra  y github.com/spf13/viper se fue capaz de desarrollar una interfaz  de línea de comandos para poder usar el proyecto de una manera mas facil. Permitiendo crear un comando por cada una de las funcionalidades que se desea usar. Los comandos son:

- serverStart
 Este comando inmediatamente levanta un servidor para la aplicacion

- addFile filepath tag1 tag2 ...tagn
  Este comando inmendiatamente añade un archivo al sistema, filepath es el path del archivo que se quiere
  subir y tag1, tag2 ... tagn son todas las etiquetas que se le van a asignar al archivo. Todo separado por
  espacio

- listFile tag1 tag2 ...tagn
  Este comando lista todos los archivos del sistema que posean todas las etiquetas pasadas como parametros

- addTags queryTags addTags
  Este comado añade a todos los archivos que cumplen queryTags las etiquetas en addTags.
  Es importante tener en cuenta que queryTags tiene la forma tag1-tag2-tagn. O sea la lista de  etiquetas estan separadas por '-'. Lo mismo pasa para addTags

- deleteTags queryTags deleteTags
  Este comado elimina de todos los archivos que cumplen queryTags las etiquetas en deleteTags.
  Es importante tener en cuenta que queryTags tiene la forma tag1-tag2-tagn. O sea la lista de etiquetas estan separadas por '-'. Lo mismo pasa para deleteTags

Es importante tener en cuenta que antes de usar el proyecto se debe crear una red de prueba usando el comando de Docker

```powershell
sudo docker network create tagNet
```

Ademas para ejecutar cada comando del proyecto se debe cargar la imagen. Por lo tanto se debe llamar de la  siguiente forma.

```powershell
sudo docker run --rm -it --network tagNet --name test1 tagfile:v1 Comando
```
