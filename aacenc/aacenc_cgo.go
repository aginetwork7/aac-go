//go:build !ignore

package aacenc

/*
#cgo linux CFLAGS: -std=gnu99 -Iexternal/aacenc/include -I/opt/armv7l-linux-musleabihf-cross/armv7l-linux-musleabihf/include -DUSE_DEFAULT_MEM -DARMV5E -DARMV7Neon -DARM_INASM -DARMV5_INASM -DARMV6_INASM -march=armv7-a -mfpu=neon -Wall
#cgo linux LDFLAGS: -mfpu=neon -L/opt/armv7l-linux-musleabihf-cross/armv7l-linux-musleabihf/lib
#cgo darwin CFLAGS: -std=gnu99 -Iexternal/aacenc/include -DUSE_DEFAULT_MEM -Wall
#include "external/aacenc/src/cmnMemory.c"
#include "external/aacenc/src/basicop2.c"
#include "external/aacenc/src/oper_32b.c"
#include "external/aacenc/src/aac_rom.c"
#include "external/aacenc/src/aacenc.c"
#include "external/aacenc/src/aacenc_core.c"
#include "external/aacenc/src/adj_thr.c"
#include "external/aacenc/src/band_nrg.c"
#include "external/aacenc/src/bit_cnt.c"
#include "external/aacenc/src/bitbuffer.c"
#include "external/aacenc/src/bitenc.c"
#include "external/aacenc/src/block_switch.c"
#include "external/aacenc/src/channel_map.c"
#include "external/aacenc/src/dyn_bits.c"
#include "external/aacenc/src/grp_data.c"
#include "external/aacenc/src/interface.c"
#include "external/aacenc/src/line_pe.c"
#include "external/aacenc/src/memalign.c"
#include "external/aacenc/src/ms_stereo.c"
#include "external/aacenc/src/pre_echo_control.c"
#include "external/aacenc/src/psy_configuration.c"
#include "external/aacenc/src/psy_main.c"
#include "external/aacenc/src/qc_main.c"
#include "external/aacenc/src/quantize.c"
#include "external/aacenc/src/sf_estim.c"
#include "external/aacenc/src/spreading.c"
#include "external/aacenc/src/stat_bits.c"
#include "external/aacenc/src/tns.c"
#include "external/aacenc/src/transform.c"
#ifdef __linux__
	#include "external/aacenc/src/AutoCorrelation_v5.s"
	#include "external/aacenc/src/CalcWindowEnergy_v5.s"
	#include "external/aacenc/src/band_nrg_v5.s"
	#include "external/aacenc/src/PrePostMDCT_v7.s"
	#include "external/aacenc/src/R4R8First_v7.s"
	#include "external/aacenc/src/Radix4FFT_v7.s"
#endif
*/
import "C"
