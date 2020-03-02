/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-X2-IEs"
 * 	found in "spec/e2sm-gNB-X2-v308.asn1"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#ifndef	_E2_Criticality_H_
#define	_E2_Criticality_H_


#include "asn_application.h"

/* Including external dependencies */
#include "NativeEnumerated.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Dependencies */
typedef enum E2_Criticality {
	E2_Criticality_reject	= 0,
	E2_Criticality_ignore	= 1,
	E2_Criticality_notify	= 2
} e_E2_Criticality;

/* E2_Criticality */
typedef long	 E2_Criticality_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_E2_Criticality;
asn_struct_free_f E2_Criticality_free;
asn_struct_print_f E2_Criticality_print;
asn_constr_check_f E2_Criticality_constraint;
ber_type_decoder_f E2_Criticality_decode_ber;
der_type_encoder_f E2_Criticality_encode_der;
xer_type_decoder_f E2_Criticality_decode_xer;
xer_type_encoder_f E2_Criticality_encode_xer;
per_type_decoder_f E2_Criticality_decode_uper;
per_type_encoder_f E2_Criticality_encode_uper;
per_type_decoder_f E2_Criticality_decode_aper;
per_type_encoder_f E2_Criticality_encode_aper;

#ifdef __cplusplus
}
#endif

#endif	/* _E2_Criticality_H_ */
#include "asn_internal.h"
