## Notes and Assumptions

Se va a crear una app de administración de features flags

La información se va a guardar en una base de datos relacional

Como las consultas van a ser muy frecuentes pero la escritura no, se va a crear un espejo, a modo de cache, de la base de datos en memoria (posiblemente redis) para mejorar el rendimiento de las consultas

El cache se va a actalizar cada vez que se haga un cambio en la base de datos, para mantener la consistencia entre ambos

Van a existir mecanismos de actualización de cache manual

se espera poder definir los siguientes datos:
* feature name (unico e identificador)
* feature description
* feature status (open, closed, whitelisted)
* status date (timestamp del ultimo cambio de estado)
* Whitelist of users (lista de usuarios que pueden usar la feature)

Se van a desarrollar endpoints de gestion CRUD para las features (API Internal), junto con el mecanismo de actualización de cache manual.

Se va a desarrollar un endpoint para consultar el estado de una feature específica para un usuario específico, se va leer desde el caché (API External)

se va a desarrollar una pequeña aplicación para poder administrar las features. (Backoffice)

No se va a implementar seguridad en este alcance, se asume que el backoffice queda dentro de una red segura, y que la API externa va a ser consumida por un BFF que ya va a estar asegurado.

