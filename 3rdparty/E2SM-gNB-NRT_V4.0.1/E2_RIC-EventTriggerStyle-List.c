/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-gNB-NRT-IEs"
 * 	found in "spec/e2sm-gNB-NRT-v401.asn"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example`
 */

#include "E2_RIC-EventTriggerStyle-List.h"

asn_TYPE_member_t asn_MBR_E2_RIC_EventTriggerStyle_List_1[] = {
	{ ATF_NOFLAGS, 0, offsetof(struct E2_RIC_EventTriggerStyle_List, ric_EventTriggerStyle_Type),
		(ASN_TAG_CLASS_CONTEXT | (0 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_E2_RIC_Style_Type,
		0,
		{ 0, 0, 0 },
		0, 0, /* No default value */
		"ric-EventTriggerStyle-Type"
		},
	{ ATF_NOFLAGS, 0, offsetof(struct E2_RIC_EventTriggerStyle_List, ric_EventTriggerStyle_Name),
		(ASN_TAG_CLASS_CONTEXT | (1 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_E2_RIC_Style_Name,
		0,
		{ 0, 0, 0 },
		0, 0, /* No default value */
		"ric-EventTriggerStyle-Name"
		},
	{ ATF_NOFLAGS, 0, offsetof(struct E2_RIC_EventTriggerStyle_List, ric_EventTriggerFormat_Type),
		(ASN_TAG_CLASS_CONTEXT | (2 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_E2_RIC_Format_Type,
		0,
		{ 0, 0, 0 },
		0, 0, /* No default value */
		"ric-EventTriggerFormat-Type"
		},
};
static const ber_tlv_tag_t asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1[] = {
	(ASN_TAG_CLASS_UNIVERSAL | (16 << 2))
};
static const asn_TYPE_tag2member_t asn_MAP_E2_RIC_EventTriggerStyle_List_tag2el_1[] = {
    { (ASN_TAG_CLASS_CONTEXT | (0 << 2)), 0, 0, 0 }, /* ric-EventTriggerStyle-Type */
    { (ASN_TAG_CLASS_CONTEXT | (1 << 2)), 1, 0, 0 }, /* ric-EventTriggerStyle-Name */
    { (ASN_TAG_CLASS_CONTEXT | (2 << 2)), 2, 0, 0 } /* ric-EventTriggerFormat-Type */
};
asn_SEQUENCE_specifics_t asn_SPC_E2_RIC_EventTriggerStyle_List_specs_1 = {
	sizeof(struct E2_RIC_EventTriggerStyle_List),
	offsetof(struct E2_RIC_EventTriggerStyle_List, _asn_ctx),
	asn_MAP_E2_RIC_EventTriggerStyle_List_tag2el_1,
	3,	/* Count of tags in the map */
	0, 0, 0,	/* Optional elements (not needed) */
	3,	/* First extension addition */
};
asn_TYPE_descriptor_t asn_DEF_E2_RIC_EventTriggerStyle_List = {
	"RIC-EventTriggerStyle-List",
	"RIC-EventTriggerStyle-List",
	&asn_OP_SEQUENCE,
	asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1,
	sizeof(asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1)
		/sizeof(asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1[0]), /* 1 */
	asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1,	/* Same as above */
	sizeof(asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1)
		/sizeof(asn_DEF_E2_RIC_EventTriggerStyle_List_tags_1[0]), /* 1 */
	{ 0, 0, SEQUENCE_constraint },
	asn_MBR_E2_RIC_EventTriggerStyle_List_1,
	3,	/* Elements count */
	&asn_SPC_E2_RIC_EventTriggerStyle_List_specs_1	/* Additional specs */
};

