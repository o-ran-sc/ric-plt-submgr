/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-NRT-IEs"
 * 	found in "spec/e2sm-gNB-NRT-v307.asn1"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#include "E2_E2SM-gNB-NRT-EventTriggerDefinition.h"

static asn_per_constraints_t asn_PER_type_E2_E2SM_gNB_NRT_EventTriggerDefinition_constr_1 CC_NOTUSED = {
	{ APC_CONSTRAINED | APC_EXTENSIBLE,  0,  0,  0,  0 }	/* (0..0,...) */,
	{ APC_UNCONSTRAINED,	-1, -1,  0,  0 },
	0, 0	/* No PER value map */
};
static asn_TYPE_member_t asn_MBR_E2_E2SM_gNB_NRT_EventTriggerDefinition_1[] = {
	{ ATF_NOFLAGS, 0, offsetof(struct E2_E2SM_gNB_NRT_EventTriggerDefinition, choice.eventDefinition_Format1),
		(ASN_TAG_CLASS_CONTEXT | (0 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition_Format1,
		0,
		{ 0, 0, 0 },
		0, 0, /* No default value */
		"eventDefinition-Format1"
		},
};
static const asn_TYPE_tag2member_t asn_MAP_E2_E2SM_gNB_NRT_EventTriggerDefinition_tag2el_1[] = {
    { (ASN_TAG_CLASS_CONTEXT | (0 << 2)), 0, 0, 0 } /* eventDefinition-Format1 */
};
static asn_CHOICE_specifics_t asn_SPC_E2_E2SM_gNB_NRT_EventTriggerDefinition_specs_1 = {
	sizeof(struct E2_E2SM_gNB_NRT_EventTriggerDefinition),
	offsetof(struct E2_E2SM_gNB_NRT_EventTriggerDefinition, _asn_ctx),
	offsetof(struct E2_E2SM_gNB_NRT_EventTriggerDefinition, present),
	sizeof(((struct E2_E2SM_gNB_NRT_EventTriggerDefinition *)0)->present),
	asn_MAP_E2_E2SM_gNB_NRT_EventTriggerDefinition_tag2el_1,
	1,	/* Count of tags in the map */
	0, 0,
	1	/* Extensions start */
};
asn_TYPE_descriptor_t asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition = {
	"E2SM-gNB-NRT-EventTriggerDefinition",
	"E2SM-gNB-NRT-EventTriggerDefinition",
	&asn_OP_CHOICE,
	0,	/* No effective tags (pointer) */
	0,	/* No effective tags (count) */
	0,	/* No tags (pointer) */
	0,	/* No tags (count) */
	{ 0, &asn_PER_type_E2_E2SM_gNB_NRT_EventTriggerDefinition_constr_1, CHOICE_constraint },
	asn_MBR_E2_E2SM_gNB_NRT_EventTriggerDefinition_1,
	1,	/* Elements count */
	&asn_SPC_E2_E2SM_gNB_NRT_EventTriggerDefinition_specs_1	/* Additional specs */
};

