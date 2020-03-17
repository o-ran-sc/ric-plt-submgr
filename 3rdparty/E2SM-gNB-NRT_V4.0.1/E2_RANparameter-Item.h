/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-NRT-IEs"
 * 	found in "spec/e2sm-gNB-NRT-v401.asn"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#ifndef	_E2_RANparameter_Item_H_
#define	_E2_RANparameter_Item_H_


#include "asn_application.h"

/* Including external dependencies */
#include "E2_RANparameter-ID.h"
#include "E2_RANparameter-Value.h"
#include "constr_SEQUENCE.h"

#ifdef __cplusplus
extern "C" {
#endif

/* E2_RANparameter-Item */
typedef struct E2_RANparameter_Item {
	E2_RANparameter_ID_t	 ranParameter_ID;
	E2_RANparameter_Value_t	 ranParameter_Value;
	/*
	 * This type is extensible,
	 * possible extensions are below.
	 */
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} E2_RANparameter_Item_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_E2_RANparameter_Item;
extern asn_SEQUENCE_specifics_t asn_SPC_E2_RANparameter_Item_specs_1;
extern asn_TYPE_member_t asn_MBR_E2_RANparameter_Item_1[2];

#ifdef __cplusplus
}
#endif

#endif	/* _E2_RANparameter_Item_H_ */
#include "asn_internal.h"
