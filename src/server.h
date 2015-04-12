/* 
 * File:   server.h
 * Author: sapk
 *
 * Created on 19 décembre 2014, 18:17
 */


#ifndef SERVER_H
#define	SERVER_H

#ifdef	__cplusplus
extern "C" {
#endif
#include <onion/onion.h>
#include <onion/exportlocal.h>
#include <onion/auth_pam.h>

    int server(int argc, char **argv);

#ifdef	__cplusplus
}
#endif

#endif	/* TILE_H */

