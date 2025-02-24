#include <stdio.h>
#include <stdlib.h>
#include <openssl/rsa.h>
#include <openssl/err.h>
#include <openssl/rand.h>
#include <getopt.h>

void handle_error(const char *msg) {
    fprintf(stderr, "%s\n", msg);
    ERR_print_errors_fp(stderr);
    exit(EXIT_FAILURE);
}

int main(int argc, char *argv[]) {
    int key_size = 0; // Default key size
    int opt;

    // Parse command-line options
    while ((opt = getopt(argc, argv, "s:")) != -1) {
        switch (opt) {
            case 's':
                key_size = atoi(optarg);
                if (key_size < 512) {
                    fprintf(stderr, "Key size must be at least 512 bits.\n");
                    exit(EXIT_FAILURE);
                }
                break;
            default:
                fprintf(stderr, "Usage: %s [-s key_size]\n", argv[0]);
                exit(EXIT_FAILURE);
        }
    }

    // Initialize OpenSSL
    OpenSSL_add_all_algorithms();
    ERR_load_crypto_strings();

    // Generate RSA key
    RSA *rsa_key = RSA_new();
    BIGNUM *e = BN_new();
    BN_set_word(e, RSA_F4); // RSA_F4 is the typical exponent value (65537)

    if (RSA_generate_key_ex(rsa_key, key_size, e, NULL) != 1) {
        handle_error("Failed to generate RSA key");
    }

    // Clean up
    RSA_free(rsa_key);
    BN_free(e);
    EVP_cleanup();
    ERR_free_strings();
    return 0;
}