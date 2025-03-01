/* The Computer Language Benchmarks Game
* https://salsa.debian.org/benchmarksgame-team/benchmarksgame/
*
* contributed by Ledrug Katz
*
*/

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

/* this depends highly on the platform.  It might be faster to use
   char type on 32-bit systems; it might be faster to use unsigned. */


typedef int elem;

typedef struct {
   elem s[16];
   elem t[16];
   int maxflips;
   int max_n;
   int odd;
   int checksum;
} pfannkuch;

int flip(pfannkuch* pf)
{
   register int i;
   register elem *x, *y, c;

   for (x = pf->t, y = pf->s, i = pf->max_n; i--; )
      *x++ = *y++;
   i = 1;
   do {
      for (x = pf->t, y = pf->t + pf->t[0]; x < y; )
         c = *x, *x++ = *y, *y-- = c;
      i++;
   } while (pf->t[pf->t[0]]);
   return i;
}

void rotate(pfannkuch* pf, int n)
{
   elem c;
   register int i;
   c = pf->s[0];
   for (i = 1; i <= n; i++) pf->s[i-1] = pf->s[i];
   pf->s[n] = c;
}

/* Tompkin-Paige iterative perm generation */
void tk(pfannkuch* pf)
{
   int i = 0, f;
   elem c[16] = {0};
   int n = pf->max_n;
   while (i < n) {
      rotate(pf, i);
      if (c[i] >= i) {
         c[i++] = 0;
         continue;
      }

      c[i]++;
      i = 1;
      pf->odd = ~pf->odd;
      if (*pf->s) {
         f = pf->s[pf->s[0]] ? flip(pf) : 1;
         if (f > pf->maxflips) pf->maxflips = f;
         pf->checksum += pf->odd ? -f : f;
      }
   }
}

int main(int argc, char **v)
{
   int i;

   if (argc < 2) {
      fprintf(stderr, "usage: %s number\n", v[0]);
      exit(1);
   }
   pfannkuch pf = {};
   pf.max_n = atoi(v[1]);
   if ( pf.max_n < 3 || pf.max_n > 15) {
      fprintf(stderr, "range: must be 3 <= n <= 12\n");
      exit(1);
   }

   for (i = 0; i < pf.max_n; i++) pf.s[i] = i;
   tk(&pf);

   printf("%d\nPfannkuchen(%d) = %d\n", pf.checksum, pf.max_n, pf.maxflips);

   return 0;
}
