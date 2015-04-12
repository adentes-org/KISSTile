
#include <onion/onion.h>
#include <onion/exportlocal.h>
#include <onion/auth_pam.h>

int server(int argc, char **argv) {
    onion *o = onion_new(O_THREADED);
//    onion_set_certificate(o, O_SSL_CERTIFICATE_KEY, "cert.pem", "cert.key", O_SSL_NONE);
    //onion_set_root_handler(o, onion_handler_auth_pam("Onion Example", "login", onion_handler_export_local_new(".")));
    onion_set_root_handler(o, onion_handler_export_local_new("."));
    
    onion_listen(o);
    onion_free(o);
    return 0;
}