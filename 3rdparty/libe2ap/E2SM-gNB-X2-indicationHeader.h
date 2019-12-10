/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-X2-IEs"
 * 	found in "Spec/e2_and_x2-combined-and-minimized.asn1"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#ifndef	_E2SM_gNB_X2_indicationHeader_H_
#define	_E2SM_gNB_X2_indicationHeader_H_


#include "asn_application.h"

/* Including external dependencies */
#include "Interface-ID.h"
#include "InterfaceDirection.h"
#include "TimeStamp.h"
#include "constr_SEQUENCE.h"

#ifdef __cplusplus
extern "C" {
#endif

/* E2SM-gNB-X2-indicationHeader */
typedef struct E2SM_gNB_X2_indicationHeader {
	Interface_ID_t	 interface_ID;
	InterfaceDirection_t	 interfaceDirection;
	TimeStamp_t	*timestamp;	/* OPTIONAL */
	/*
	 * This type is extensible,
	 * possible extensions are below.
	 */
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} E2SM_gNB_X2_indicationHeader_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_E2SM_gNB_X2_indicationHeader;

#ifdef __cplusplus
}
#endif

#endif	/* _E2SM_gNB_X2_indicationHeader_H_ */
#include "asn_internal.h"
