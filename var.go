package go9p

import "flag"

/* enables "Akaros" capabilities, which right now means
 * a sane error message format.
 */
var Akaros = flag.Bool("akaros", false, "Akaros extensions")
