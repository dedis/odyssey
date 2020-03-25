# Attributes

We identified three main type of attributes that the data owners should be able
to set in order to gain full control of their datasets:

- **purpose**
- **classification**
- **access**

The attributes can be described in json syntax, which is convenient to update
the catalog. The attributes are split into two categories: must_have and
allowed. A "must_have" attribute is an attribute that must be selected by the
data scientist to be accepted. This is the case of the "access" attributes. An
"allowed" attribute is an attribute that can be selected by the data scientist
but is not compulsory. This is the case of the "purpose" and "classification"
attributes.

An attribute can have a "delegated_enforcement", which is necessary for
attributes that can not be automatically validated because there is a textual
description that a data scientist must read and agree on. Those textual
descriptions are written in a special "delegated_enforcement" section, where we
link delegated attributes and the textual description with their IDs.

Finally, we designed attributes to be recursively defined, ie. an attribute can
have multiple attributes, which can have multiple attributes and so on. This
gives us a flexible way to define complex attributes structures.

## Definition


```json
{
	"attributesGroups": [{
			"title": "Use",
			"description": "How this dataset can be used",
			"consumer_description": "Please tell us how the result will be used",
			"attributes": [{
					"id": "use_restricted",
					"name": "use_restricted",
					"description": "The use of this dataset is restricted",
					"type": "checkbox",
					"rule_type": "must_have",
					"delegated_enforcement": true,
					"attributes": [{
						"id": "use_restricted_description",
						"name": "use_restricted_description",
						"description": "Please describe the restriction",
						"type": "text",
						"rule_type": "must_have",
						"delegated_enforcement": true,
						"attributes": []
					}]
				},
				{
					"id": "use_retention_policy",
					"name": "use_retention_policy",
					"description": "There is a special retention policy for this dataset",
					"type": "checkbox",
					"rule_type": "must_have",
					"delegated_enforcement": true, 
					"attributes": [{
						"id": "use_retention_policy_description",
						"name": "use_retention_policy_description",
						"description": "Please describe the retention policy",
						"type": "text",
						"rule_type": "must_have",
						"delegated_enforcement": true,
						"attributes": []
					}]
				},
				{
					"id": "use_predefined_purpose",
					"name": "use_predefined_purpose",
					"description": "Can be used for the following predefined-purposess",
					"consumer_description": "the result will be used for the following pre-defined purposes",
					"type": "checkbox",
					"rule_type": "allowed",
					"attributes": [{
						"id": "use_predefined_purpose_legal",
						"name": "use_predefined_purpose_legal",
						"description": "To meet company's legal or regulatory requirements",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_analytics_counterparty",
						"name": "use_predefined_purpose_analytics_counterparty",
						"description": "To provide analytics to the counterparty of the contract from which data is sourced/for the counterparty's benefit",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_essential",
						"name": "use_predefined_purpose_essential",
						"description": "To perform an essential function of company's business (for example, reserving, develop and improve costing/pricing models, portfolio management, enable risk modelling and accumulation control)",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_analytics_multiple_counterparty",
						"name": "use_predefined_purpose_analytics_multiple_counterparty",
						"description": "To provide analytics to multiple counterparties / for the benefit of multiple counterparties",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_general_information",
						"name": "use_predefined_purpose_general_information",
						"description": "To provide general information to the public",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_sole_company",
						"name": "use_predefined_purpose_sole_company",
						"description": "For the sole benefit of the company",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}]
				}
			]
		},
		{
			"title": "Classification",
			"description": "The following classification types apply on my dataset",
			"consumer_description": "I have the right to work with the following types of data",
			"attributes": [{
				"id": "classification_public",
				"name": "classification_public",
				"description": "Public information",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_internal",
				"name": "classification_internal",
				"description": "Internal data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_confidential",
				"name": "classification_confidential",
				"description": "Confidential data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_critical",
				"name": "classification_critical",
				"description": "Critical data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_personal",
				"name": "classification_personal",
				"description": "Personal data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}]
		},
		{
			"title": "Access",
			"description": "Tell us who can access the data",
			"consumer_description": "Please select who will have access the the result",
			"attributes": [{
				"id": "access_unrestricted",
				"name": "access",
				"description": "No restriction",
				"type": "radio",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "access_internal",
				"name": "access",
				"description": "internal",
				"type": "radio",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "access_defined_group",
				"name": "access",
				"description": "defined group",
				"type": "radio",
				"rule_type": "must_have",
				"attributes": [{
					"id": "access_defined_group_description",
					"name": "access_defined_group_description",
					"description": "Please specify the group",
					"delegated_enforcement": true,
					"type": "text",
					"rule_type": "must_have"
				}]
			}]
		}
	],
	"delegated_enforcement": {
		"title": "Manual enforcement",
		"description": "some restriction can not be automatically checked. Therefore, you are requested to agree on the following attributes.",
		"attributes": [{
			"id": "use_restricted_description_enforcement",
			"description": "I agree on this restriction use",
			"value_from_id": "use_restricted_description",
			"trigger_id": "use_restricted",
			"trigger_value": "",
			"check_validates": [
				"use_restricted"
			],
			"text_validates": "use_restricted_description"
		},{
			"id": "use_retention_policy_description_enforcement",
			"description": "I agree on this retention policy",
			"value_from_id": "use_retention_policy_description",
			"trigger_id": "use_retention_policy",
			"trigger_value": "",
			"check_validates": [
				"use_retention_policy"
			],
			"text_validates": "use_retention_policy_description"
		},{
			"id": "access_defined_group_description_enforcement",
			"description": "I certify that the result will only be used by this specific group",
			"value_from_id": "access_defined_group_description",
			"trigger_id": "access_defined_group",
			"trigger_value": "",
			"text_validates": "access_defined_group_description"
		}]
	}
}
```

## DARC

Translated into DARC custom attributes, a data owner can express the three types
of attributes with the following rules:

```
Darc(data owner)
    spawn:calypsoread - 
		attr:allowed:
			use_restricted=checked&
			use_restricted_description_29e58702ba0524ef9eac162914016241f795137aef54a2670979e887925ed9fa=This+is+the+description+of+the+restriction&use_retention_policy=checked&
			use_retention_policy_description_29e58702ba0524ef9eac162914016241f795137aef54a2670979e887925ed9fa=This+is+the+retention+policy&classification_public=checked&
			classification_internal=checked&classification_confidential=checked&
			classification_critical=checked&classification_personal=checked&
			access_defined_group=checked&
			access_defined_group_description_29e58702ba0524ef9eac162914016241f795137aef54a2670979e887925ed9fa=This+is+the+specific+group+description&use_predefined_purpose=checked&
			use_predefined_purpose_legal=checked&use_predefined_purpose_analytics_counterparty=checked&use_predefined_purpose_essential=checked&
			use_predefined_purpose_analytics_multiple_counterparty=checked&
			use_predefined_purpose_general_information=checked&
			use_predefined_purpose_sole_company=checked&
		attr:must_have:
			use_restricted=checked&
			use_restricted_description_29e58702ba0524ef9eac162914016241f795137aef54a2670979e887925ed9fa=This+is+the+description+of+the+restriction&use_retention_policy=checked&
			use_retention_policy_description_29e58702ba0524ef9eac162914016241f795137aef54a2670979e887925ed9fa=This+is+the+retention+policy&
			classification_public=checked&
			classification_internal=checked&
			classification_confidential=checked&
			classification_critical=checked&
			classification_personal=checked&
			access_defined_group=checked&
			access_defined_group_description_29e58702ba0524ef9eac162914016241f795137aef54a2670979e887925ed9fa=This+is+the+specific+group+description
```