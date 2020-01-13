# Attributes

We identified three main type of attributes that the data owners should be able
to set in order to gain full control of their datasets:

- **purpose**
- **classification**
- **access**

## Definition


```json
{
	"attributesGroups": [{
			"title": "Use",
			"description": "How this dataset can be used",
			"attributes": [{
					"id": "use_restricted",
					"description": "The use of this dataset is restricted",
					"type": "checkbox",
					"attributes": [{
						"id": "use_restricted_description",
						"description": "Please describe the restriction",
						"type": "text",
						"attributes": []
					}]
				},
				{
					"id": "use_retention_policy",
					"description": "There is a special retention policy for this dataset",
					"type": "checkbox",
					"attributes": [{
						"id": "use_retention_policy_description",
						"description": "Please describe the retention policy",
						"type": "text",
						"attributes": []
					}]
				},
				{
					"id": "use_predefined_purpose",
					"description": "Can be used for the following predefined-purposess",
					"type": "checkbox",
					"attributes": [{
						"id": "use_predefined_purpose_legal",
						"description": "To meet legal or regulatory requirements",
						"type": "checkbox",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_analytics_counterparty",
						"description": "To provide analytics to the counterparty of the contract from which data is sourced/for the counterparty's benefit",
						"type": "checkbox",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_essential",
						"description": "To perform an essential function of business (for example, reserving, develop and improve costing/pricing models, portfolio management, enable risk modelling and accumulation control)",
						"type": "checkbox",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_analytics_multiple_counterparty",
						"description": "To provide analytics to multiple counterparties / for the benefit of multiple counterparties",
						"type": "checkbox",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_general_information",
						"description": "To provide general information to the public",
						"type": "checkbox",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_sole_sr",
						"description": "For the sole benefit of the company",
						"type": "checkbox",
						"attributes": []
					}]
				}
			]
		},
		{
			"title": "Classification",
			"description": "The following classification types apply on my dataset",
			"attributes": [{
				"id": "classification_public",
				"description": "Public information",
				"type": "checkbox",
				"attributes": []
			}, {
				"id": "classification_internal",
				"description": "Internal data",
				"type": "checkbox",
				"attributes": []
			}, {
				"id": "classification_confidential",
				"description": "Confidential data",
				"type": "checkbox",
				"attributes": []
			}, {
				"id": "classification_critical",
				"description": "Critical data",
				"type": "checkbox",
				"attributes": []
			}, {
				"id": "classification_personal",
				"description": "Personal data",
				"type": "checkbox",
				"attributes": []
			}]
		},
		{
			"title": "Access",
			"description": "Tell us who can access the data",
			"attributes": [{
				"id": "access_unrestricted",
				"description": "No restriction",
				"type": "radio",
				"name": "access",
				"attributes": []
			}, {
				"id": "access_internal",
				"description": "SR internal",
				"type": "radio",
				"name": "access",
				"attributes": []
			}, {
				"id": "access_defined_group",
				"description": "SR defined group",
				"type": "radio",
				"name": "access",
				"attributes": [{
					"id": "access_defined_group_description",
					"description": "Please specify the group",
					"type": "text"
				}]
			}]
		}
	]
}
```

## DARC

Translated into DARC custom attributes, a data owner can express the three types
of attributes with the following rules:

```
Darc(data owner)
    spawn:calypsoread - 
		attr:allowed:use_restricted=checked&
			use_restricted_description_4c5870a8302c88283791bc1467c2b1acdd2ba84fc8ea7c57d6cc8947798fc4b7=Can+not+be+used+when+it+is+raining+outside&use_retention_policy=checked&
			use_retention_policy_description_4c5870a8302c88283791bc1467c2b1acdd2ba84fc8ea7c57d6cc8947798fc4b7=This+is+my+special+retention+policy&
			use_predefined_purpose=checked&
			use_predefined_purpose_legal=checked&
			use_predefined_purpose_analytics_counterparty=checked&
			use_predefined_purpose_essential=checked&use_predefined_purpose_analytics_multiple_counterparty=checked&use_predefined_purpose_general_information=checked&
			use_predefined_purpose_sole_sr=checked&
			classification_public=checked&
			classification_internal=checked&
			classification_confidential=checked&
			classification_critical=checked&
			classification_personal=checked&
			access_defined_group=checked&access_defined_group_description_4c5870a8302c88283791bc1467c2b1acdd2ba84fc8ea7c57d6cc8947798fc4b7=Only+the+legal+group+can+use+this+dataset&
```