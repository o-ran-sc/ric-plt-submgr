/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-X2-IEs"
 * 	found in "spec/e2sm-gNB-X2-v308.asn1"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#ifndef	_E2_InterfaceProtocolIE_ID_H_
#define	_E2_InterfaceProtocolIE_ID_H_


#include "asn_application.h"

/* Including external dependencies */
#include "E2_ProtocolIE-ID.h"

#ifdef __cplusplus
extern "C" {
#endif

/* E2_InterfaceProtocolIE-ID */
typedef E2_ProtocolIE_ID_t	 E2_InterfaceProtocolIE_ID_t;

/* Implementation */
extern asn_per_constraints_t asn_PER_type_E2_InterfaceProtocolIE_ID_constr_1;
extern asn_TYPE_descriptor_t asn_DEF_E2_InterfaceProtocolIE_ID;
asn_struct_free_f E2_InterfaceProtocolIE_ID_free;
asn_struct_print_f E2_InterfaceProtocolIE_ID_print;
asn_constr_check_f E2_InterfaceProtocolIE_ID_constraint;
ber_type_decoder_f E2_InterfaceProtocolIE_ID_decode_ber;
der_type_encoder_f E2_InterfaceProtocolIE_ID_encode_der;
xer_type_decoder_f E2_InterfaceProtocolIE_ID_decode_xer;
xer_type_encoder_f E2_InterfaceProtocolIE_ID_encode_xer;
per_type_decoder_f E2_InterfaceProtocolIE_ID_decode_uper;
per_type_encoder_f E2_InterfaceProtocolIE_ID_encode_uper;
per_type_decoder_f E2_InterfaceProtocolIE_ID_decode_aper;
per_type_encoder_f E2_InterfaceProtocolIE_ID_encode_aper;

#ifdef __cplusplus
}
#endif

#endif	/* _E2_InterfaceProtocolIE_ID_H_ */
#include "asn_internal.h"
