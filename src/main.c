/* 
 * File:   main.c
 * Author: sapk
 *
 * Created on 19 décembre 2014, 18:08
 */

#include <stdio.h>
#include <stdlib.h>
#include <onion/onion.h>
#include <onion/exportlocal.h>
#include <onion/auth_pam.h>
#include "server.h"

/*
 * 
 */
int main(int argc, char** argv) {

    printf("Hello World !");
    
    return(server(argc,argv));
//    return (EXIT_SUCCESS);
}

