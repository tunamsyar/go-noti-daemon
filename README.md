# Matt Daemon

A simple webserver that receives a json body to send to FCM.

Attempted to daemonize the server so it could run only on the binary.

Probably some hardening is needed to ensure if the endpoint is leaked,
it won't be bombarded.

