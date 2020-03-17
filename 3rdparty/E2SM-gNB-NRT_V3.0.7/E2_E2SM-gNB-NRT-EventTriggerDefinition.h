/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-NRT-IEs"
 * 	found in "spec/e2sm-gNB-NRT-v307.asn1"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#ifndef	_E2_E2SM_gNB_NRT_EventTriggerDefinition_H_
#define	_E2_E2SM_gNB_NRT_EventTriggerDefinition_H_


#include "asn_application.h"

/* Including external dependencies */
#include "E2_E2SM-gNB-NRT-EventTriggerDefinition-Format1.h"
#include "constr_CHOICE.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Dependencies */
typedef enum E2_E2SM_gNB_NRT_EventTriggerDefinition_PR {
	E2_E2SM_gNB_NRT_EventTriggerDefinition_PR_NOTHING,	/* No components present */
	E2_E2SM_gNB_NRT_EventTriggerDefinition_PR_eventDefinition_Format1
	/* Extensions may appear below */
	
} E2_E2SM_gNB_NRT_EventTriggerDefinition_PR;

/* E2_E2SM-gNB-NRT-EventTriggerDefinition */
typedef struct E2_E2SM_gNB_NRT_EventTriggerDefinition {
	E2_E2SM_gNB_NRT_EventTriggerDefinition_PR present;
	union E2_E2SM_gNB_NRT_EventTriggerDefinition_u {
		E2_E2SM_gNB_NRT_EventTriggerDefinition_Format1_t	 eventDefinition_Format1;
		/*
		 * This type is extensible,
		 * possible extensions are below.
		 */
	} choice;
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} E2_E2SM_gNB_NRT_EventTriggerDefinition_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition;

#ifdef __cplusplus
}
#endif

#endif	/* _E2_E2SM_gNB_NRT_EventTriggerDefinition_H_ */
#include "asn_internal.h"
