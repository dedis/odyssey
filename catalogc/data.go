package catalogc

import (
	"fmt"
	"net/url"
	"strings"

	"go.dedis.ch/cothority/v3/byzcoin"
	"golang.org/x/xerrors"
)

// CatalogData holds the data of the Odyssey catalog contract
type CatalogData struct {
	Owners   []*Owner
	Metadata *Metadata
}

// String returns a human readable string representation of the project data
func (cd CatalogData) String() string {
	out := new(strings.Builder)
	out.WriteString("- Catalog:\n")
	out.WriteString("-- Owners:\n")
	for i, owner := range cd.Owners {
		fmt.Fprintf(out, "--- Owners[%d]:\n", i)
		out.WriteString(eachLine.ReplaceAllString(owner.String(), "---$1"))
	}
	if cd.Metadata != nil {
		out.WriteString(eachLine.ReplaceAllString(cd.Metadata.String(), "-$1"))
	}
	return out.String()
}

// GetOwner returns the owner if found, or nil
func (cd CatalogData) GetOwner(identityStr string) *Owner {
	for _, owner := range cd.Owners {
		if owner.IdentityStr == identityStr {
			return owner
		}
	}
	return nil
}

// AddOwner adds a new owner if not already present
func (cd *CatalogData) AddOwner(owner *Owner) error {
	foundOwner := cd.GetOwner(owner.IdentityStr)
	if foundOwner != nil {
		return xerrors.Errorf("an owner with identityStr '%s' is already "+
			"present:\n%s", owner.IdentityStr, owner.String())
	}
	if cd.Owners == nil {
		cd.Owners = make([]*Owner, 1)
	}
	cd.Owners = append(cd.Owners, owner)
	return nil
}

// RemoveOwner removes an owner from the list of owners
func (cd *CatalogData) RemoveOwner(identityStr string) error {
	index := -1
	for i, owner := range cd.Owners {
		if owner.IdentityStr == identityStr {
			index = i
			break
		}
	}
	if index == -1 {
		return xerrors.Errorf("owner with identity '%s' not found", identityStr)
	}

	// Here we replace the owner we want to delete by the last element of the
	// list, then we truncate the list by removing its last element.
	cd.Owners[index] = cd.Owners[len(cd.Owners)-1]
	cd.Owners[len(cd.Owners)-1] = nil
	cd.Owners = cd.Owners[:len(cd.Owners)-1]

	return nil
}

// Owner describes someone that has one or more datasets
type Owner struct {
	Firstname string
	Lastname  string
	Datasets  []*Dataset
	// This is the signer identity, like 'ed25519:aef123'
	IdentityStr string
}

// String returns a human readable string representation of an owner
func (o Owner) String() string {
	out := new(strings.Builder)
	out.WriteString("- Owner:\n")
	fmt.Fprintf(out, "-- Firstname: %s\n", o.Firstname)
	fmt.Fprintf(out, "-- Lastname: %s\n", o.Lastname)
	fmt.Fprintf(out, "-- IdentityStr: %s\n", o.IdentityStr)
	out.WriteString("-- Datasets:\n")
	for i, dataset := range o.Datasets {
		fmt.Fprintf(out, "--- Datasets[%d]:\n", i)
		out.WriteString(eachLine.ReplaceAllString(dataset.String(), "---$1"))
	}
	return out.String()
}

// GetDataset return the dataset if found, or nil
func (o Owner) GetDataset(calypsoWriteID string) *Dataset {
	for _, dataset := range o.Datasets {
		if dataset.CalypsoWriteID == calypsoWriteID {
			return dataset
		}
	}
	return nil
}

// AddDataset adds a dataset if it doesn't already exist
func (o *Owner) AddDataset(dataset *Dataset) error {
	foundDataset := o.GetDataset(dataset.CalypsoWriteID)
	if foundDataset != nil {
		return xerrors.Errorf("dataset with calypsoWriteID '%s' already "+
			"exist:\n%s", dataset.CalypsoWriteID, dataset.String())
	}
	if o.Datasets == nil {
		o.Datasets = make([]*Dataset, 1)
	}
	o.Datasets = append(o.Datasets, dataset)
	return nil
}

// ReplaceDataset replace the dataset at the given calypsoWriteID by the
// newDataset
func (o *Owner) ReplaceDataset(calypsoWriteID string, newDataset *Dataset) error {
	index := -1
	for i, dataset := range o.Datasets {
		if dataset.CalypsoWriteID == calypsoWriteID {
			index = i
			break
		}
	}
	if index == -1 {
		return xerrors.Errorf("dataset with calypsoWriteID '%s' not found",
			calypsoWriteID)
	}
	o.Datasets[index] = newDataset
	return nil
}

// DeleteDataset deletes the dataset at the given calypsoWriteID
func (o *Owner) DeleteDataset(calypsoWriteID string) error {
	index := -1
	for i, dataset := range o.Datasets {
		if dataset.CalypsoWriteID == calypsoWriteID {
			index = i
			break
		}
	}
	if index == -1 {
		return xerrors.Errorf("dataset with calypsoWriteID '%s' not found",
			calypsoWriteID)
	}

	// Here we replace the dataset we want to delete by the last element of the
	// list, then we truncate the list by removing its last element.
	o.Datasets[index] = o.Datasets[len(o.Datasets)-1]
	o.Datasets[len(o.Datasets)-1] = nil
	o.Datasets = o.Datasets[:len(o.Datasets)-1]
	return nil
}

// ArchiveDataset set the IsArchived attribute and removes its attributes
func (o *Owner) ArchiveDataset(calypsoWriteID string) error {
	index := -1
	for i, dataset := range o.Datasets {
		if dataset.CalypsoWriteID == calypsoWriteID {
			index = i
			break
		}
	}
	if index == -1 {
		return xerrors.Errorf("dataset with calypsoWriteID '%s' not found",
			calypsoWriteID)
	}

	o.Datasets[index].IsArchived = true
	// Actually I don't know if its safe to do that. We may have some surprise
	// with metadata.AttributesGroups that is nil
	o.Datasets[index].Metadata = &Metadata{}
	return nil
}

// Dataset describes the metadata of a dataset
type Dataset struct {
	CalypsoWriteID string `json:"calypsoWriteID"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	CloudURL       string `json:"cloudURL"`
	// This is the signer identity that owns the dataset, like 'ed25519:aef123'
	IdentityStr string    `json:"identityStr"`
	SHA2        string    `json:"sha2"`
	IsArchived  bool      `json:"is_archived"`
	Metadata    *Metadata `json:"metadata"`
}

// String returns a human readable string representation of a datasets
func (d Dataset) String() string {
	out := new(strings.Builder)
	out.WriteString("- Dataset:\n")
	fmt.Fprintf(out, "-- CalypsoWriteID: %s\n", d.CalypsoWriteID)
	fmt.Fprintf(out, "-- Title: %s\n", d.Title)
	fmt.Fprintf(out, "-- Description: %s\n", d.Description)
	fmt.Fprintf(out, "-- CloudURL: %s\n", d.CloudURL)
	fmt.Fprintf(out, "-- SHA2: %s\n", d.SHA2)
	fmt.Fprintf(out, "-- IdentityStr: %s\n", d.IdentityStr)
	fmt.Fprintf(out, "-- IsArchived: %v\n", d.IsArchived)
	out.WriteString("-- Metadata:\n")
	if d.Metadata != nil {
		out.WriteString(eachLine.ReplaceAllString(d.Metadata.String(), "--$1"))
	}
	return out.String()
}

// DelegatedForm prints the form where the data scientist has to agree on custum
// text attributes that can not be automatically checked. It fills the value
// attribute if it finds an attribute that has the same id in the given metadata
func (d Dataset) DelegatedForm(m *Metadata) string {
	out := new(strings.Builder)
	if d.Metadata == nil {
		out.WriteString("ERROR: METADATA IS NUL FOR DATASET " + d.Title)
		return out.String()
	}

	for i, attr := range d.Metadata.DelegatedEnforcement.Attributes {
		sourceAttr, found := d.Metadata.GetAttribute(attr.ValueFromID)
		if !found {
			out.WriteString("FROM_VALUE ATTRIBUTE WITH ID '" +
				attr.ValueFromID + "' NOT FOUND")
			continue
		}

		// All the attributes set by triggeredAttributes must have a value,
		// otherwise we do not show this delegated attribute

		triggerAttr, found := d.Metadata.GetAttribute(attr.TriggerID)
		if !found {
			out.WriteString("TRIGGERED_ATTRIBUTE WITH ID '" +
				attr.TriggerID + "' NOT FOUND")
			continue
		}
		if triggerAttr.Value == "" {
			continue
		}

		checked := ""
		_, isChecked := m.GetAttribute(attr.ValueFromID + "_" + d.CalypsoWriteID)
		if isChecked {
			checked = "checked"
		}

		fmt.Fprintf(out, "<p class='statement-title'>Statement %d for dataset '%s'</p>", i, d.Title)
		out.WriteString("<div class='statement-wrapper'>")
		fmt.Fprintf(out, "<p class='statement'>%s</p>", sourceAttr.Value)
		fmt.Fprintf(out, "<label for='%s' class='aligned-checkbox'>", attr.TextValidates+"_"+d.CalypsoWriteID)
		fmt.Fprintf(out, "<input %s required onchange=\"toggleColor(this)\" type='checkbox' id='%s' value='%s' name='%s' class='check-color'>\n", checked, attr.TextValidates+"_"+d.CalypsoWriteID, sourceAttr.Value, attr.TextValidates+"_"+d.CalypsoWriteID)
		fmt.Fprintf(out, "<span>%s</span>\n", attr.Description)
		out.WriteString("</label>")

		for _, checkValidatesID := range attr.CheckValidates {
			fmt.Fprintf(out, "<input type='hidden' name='%s' value='%s'>", checkValidatesID, "checked")
		}

		out.WriteString("</div>")
	}
	return out.String()
}

// Metadata is a struct that holds a list of attributes groups. This was
// necessary to use a struct in order to marshal/unmarshal it.
type Metadata struct {
	AttributesGroups     []*AttributesGroup    `json:"attributesGroups"`
	DelegatedEnforcement *DelegatedEnforcement `json:"delegated_enforcement"`
}

func (m Metadata) String() string {
	out := new(strings.Builder)
	out.WriteString("- Metadata:\n")
	out.WriteString("-- AttributesGroups:\n")
	for i, ag := range m.AttributesGroups {
		fmt.Fprintf(out, "--- AttributesGroups[%d]:\n", i)
		out.WriteString(eachLine.ReplaceAllString(ag.String(), "---$1"))
	}
	out.WriteString("-- DelegatedEnforcement:\n")
	if m.DelegatedEnforcement != nil {
		out.WriteString(eachLine.ReplaceAllString(m.DelegatedEnforcement.String(), "--$1"))
	}
	return out.String()
}

// TryUpdate update an attribute, if the name is found, by the value
func (m *Metadata) TryUpdate(name, val string) {
	// If there is multiple attributes with the same name, we are in the case of
	// a radio button group. In that case, we select the radio button that has
	// its ID corresponding to the value ("val"), and we set its value to
	// "checked"
	foundAttributes := m.GetAttributesByName(name)
	if len(foundAttributes) > 1 {
		for _, attr := range foundAttributes {
			if attr.ID == val {
				attr.Value = "checked"
			}
		}
		return
	}
	if len(foundAttributes) == 1 {
		foundAttributes[0].Value = val
	}
}

// UpdateOrSet update an attribute, if the name is found, by the value, or
// create a new attribute
func (m *Metadata) UpdateOrSet(name, val string) {
	// If there is multiple attributes with the same name, we are in the case of
	// a radio button group. In that case, we select the radio button that has
	// its ID corresponding to the value ("val"), and we set its value to
	// "checked"
	foundAttributes := m.GetAttributesByName(name)
	if len(foundAttributes) > 1 {
		for _, attr := range foundAttributes {
			if attr.ID == val {
				attr.Value = "checked"
			}
		}
	} else if len(foundAttributes) == 1 {
		foundAttributes[0].Value = val
	} else {
		if m.AttributesGroups == nil || len(m.AttributesGroups) == 0 {
			m.AttributesGroups = []*AttributesGroup{&AttributesGroup{}}
		}
		if m.AttributesGroups[0].Attributes == nil {
			m.AttributesGroups[0].Attributes = make([]*Attribute, 0)
		}
		m.AttributesGroups[0].Attributes = append(
			m.AttributesGroups[0].Attributes, &Attribute{
				ID: name, Name: name, Value: val})
	}
}

// FoundName checks if an attribute uses this name
func (m Metadata) FoundName(name string) bool {
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			if attr.FoundName(name) {
				return true
			}
		}
	}
	return false
}

// Reset reset the value field of all the attributes
func (m *Metadata) Reset() {
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			attr.Reset()
		}
	}
}

// Darc print the DARC representation of the metadata. Outputs something like
// "attr:allowed:ID1=value1&ID2=value2& attr:must_have:ID2=value2&ID3=value3"
func (m Metadata) Darc(calypsoWriteID string) string {
	allowedAttr := m.GetActiveAttributesByRuleType("allowed")
	mustHaveAttr := m.GetActiveAttributesByRuleType("must_have")

	outAllowed := new(strings.Builder)
	outMustHave := new(strings.Builder)

	outAllowed.WriteString("( attr:allowed:")
	outMustHave.WriteString(" & attr:must_have:")

	// A "must_have" attribute is defacto allowed
	for _, attr := range mustHaveAttr {
		outMustHave.WriteString(attr.Darc(calypsoWriteID))
		outAllowed.WriteString(attr.Darc(calypsoWriteID))
	}

	for _, attr := range allowedAttr {
		outAllowed.WriteString(attr.Darc(calypsoWriteID))
	}
	outAllowed.WriteString(outMustHave.String() + " )")
	return outAllowed.String()
}

// GetAttribute return the first attribute that has the given id, and a bool
// that says if one was found.
func (m Metadata) GetAttribute(id string) (*Attribute, bool) {
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			a, found := attr.GetAttribute(id)
			if found {
				return a, true
			}
		}
	}
	return nil, false
}

// GetAttributesByName return the attributes with the given name
func (m Metadata) GetAttributesByName(name string) []*Attribute {
	res := make([]*Attribute, 0)
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			res = append(res, attr.GetAttributesByName(name)...)
		}
	}
	return res
}

// GetAttributeByValue return the first attribute that has the given value, and
// a bool that says if one was found.
func (m Metadata) GetAttributeByValue(value string) (*Attribute, bool) {
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			a, found := attr.GetAttributeByValue(value)
			if found {
				return a, true
			}
		}
	}
	return nil, false
}

// GetActiveAttributesByRuleType return all the attributes that has the given
// rule type and a non-empty value.
func (m Metadata) GetActiveAttributesByRuleType(ruleType string) []*Attribute {
	res := make([]*Attribute, 0)
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			res = append(res, attr.GetActiveAttributesByRuleType(ruleType)...)
		}
	}
	return res
}

// ResetFailedReasons sets the FailedReasons field of every attribute to an
// empty slice
func (m *Metadata) ResetFailedReasons() {
	for _, ag := range m.AttributesGroups {
		for _, attr := range ag.Attributes {
			attr.ResetFailedReasons()
		}
	}
}

// AttributesGroup ...
type AttributesGroup struct {
	Title               string       `json:"title"`
	Description         string       `json:"description"`
	ConsumerDescription string       `json:"consumer_description"`
	Attributes          []*Attribute `json:"attributes"`
}

func (ag AttributesGroup) String() string {
	out := new(strings.Builder)
	out.WriteString("- AttributesGroup:\n")
	fmt.Fprintf(out, "-- Title: %s\n", ag.Title)
	fmt.Fprintf(out, "-- Description: %s\n", ag.Description)
	fmt.Fprintf(out, "-- ConsumerDescription: %s\n", ag.ConsumerDescription)
	out.WriteString("-- Attributes:\n")
	for i, attr := range ag.Attributes {
		fmt.Fprintf(out, "--- Attributes[%d]:\n", i)
		out.WriteString(eachLine.ReplaceAllString(attr.String(), "---$1"))
	}
	return out.String()
}

// Attribute ...
type Attribute struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Type        string `json:"type"`
	RuleType    string `json:"rule_type"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	// If true the attribute is not enforced automatically by comparing the
	// attributes in the DARC, but it uses one of the attributes in the
	// DelegatedEnforcement struct.
	DelegatedEnforcement bool           `json:"delegated_enforcement"`
	FailedReasons        *FailedReasons `json:"failed_reasons"`
	Attributes           []*Attribute   `json:"attributes"`
}

func (a Attribute) String() string {
	out := new(strings.Builder)
	out.WriteString("- Attribute:\n")
	fmt.Fprintf(out, "-- ID: %s\n", a.ID)
	fmt.Fprintf(out, "-- Description: %s\n", a.Description)
	fmt.Fprintf(out, "-- Type: %s\n", a.Type)
	fmt.Fprintf(out, "-- RuleType: %s\n", a.RuleType)
	fmt.Fprintf(out, "-- Name: %s\n", a.Name)
	fmt.Fprintf(out, "-- Value: %s\n", a.Value)
	fmt.Fprintf(out, "-- DelegatedEnforcement: %v\n", a.DelegatedEnforcement)
	if a.FailedReasons != nil {
		out.WriteString(eachLine.ReplaceAllString(a.FailedReasons.String(), "-$1"))
	}
	out.WriteString("-- Attributes:\n")
	for i, attr := range a.Attributes {
		fmt.Fprintf(out, "--- Attributes[%d]:\n", i)
		out.WriteString(eachLine.ReplaceAllString(attr.String(), "---$1"))
	}
	return out.String()
}

// TryUpdate update an attribute, if the key is found, by the value
func (a *Attribute) TryUpdate(key, val string) {
	if a.ID == key {
		a.Value = val
	}
	for _, attr := range a.Attributes {
		attr.TryUpdate(key, val)
	}
}

// FoundName checks if this attribute or its subattributes use this name
func (a Attribute) FoundName(name string) bool {
	if a.Name == name {
		return true
	}
	for _, attr := range a.Attributes {
		if attr.FoundName(name) {
			return true
		}
	}
	return false
}

// Reset reset the value field of this attribute and all its sub-attributes
func (a *Attribute) Reset() {
	a.Value = ""
	for _, attr := range a.Attributes {
		attr.Reset()
	}
}

// GetAttribute return the first attribute found that has the given id
func (a *Attribute) GetAttribute(id string) (*Attribute, bool) {
	if a.ID == id {
		return a, true
	}
	for _, attr := range a.Attributes {
		attr2, found := attr.GetAttribute(id)
		if found {
			return attr2, true
		}
	}
	return nil, false
}

// GetAttributesByName return the attributes with the given name
func (a *Attribute) GetAttributesByName(name string) []*Attribute {
	res := make([]*Attribute, 0)
	if a.Name == name {
		res = append(res, a)
	}
	for _, attr := range a.Attributes {
		res = append(res, attr.GetAttributesByName(name)...)
	}
	return res
}

// GetAttributeByValue return the first attribute found that has the given id
func (a *Attribute) GetAttributeByValue(value string) (*Attribute, bool) {
	if a.Value == value {
		return a, true
	}
	for _, attr := range a.Attributes {
		attr2, found := attr.GetAttributeByValue(value)
		if found {
			return attr2, true
		}
	}
	return nil, false
}

// GetActiveAttributesByRuleType return the attributes that have the given rule
// type and that has a value.
func (a *Attribute) GetActiveAttributesByRuleType(ruleType string) []*Attribute {
	res := make([]*Attribute, 0)
	if a.Value == "" {
		return res
	}
	if a.RuleType == ruleType {
		res = append(res, a)
	}
	for _, attr := range a.Attributes {
		res = append(res, attr.GetActiveAttributesByRuleType(ruleType)...)
	}
	return res
}

// Form print the HTML field form element of an attribute and its sub attributes
func (a Attribute) Form() string {
	out := new(strings.Builder)
	out.WriteString("<div>")
	switch a.Type {
	case "checkbox":
		fmt.Fprintf(out, "<label for='%s' class='aligned-checkbox'>", a.ID)
		// If there is a value, we check the checkbox
		if a.Value != "" {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='checkbox' id='%s' value='%s' name='%s' checked>\n", a.ID, "checked", a.Name)
		} else {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='checkbox' id='%s' value='%s' name='%s'>\n", a.ID, "checked", a.Name)
		}
		fmt.Fprintf(out, "<span>%s</span>\n", a.Description)
		out.WriteString("</label>")
	case "text":
		fmt.Fprintf(out, "<label>%s</label>\n", a.Description)
		fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" class='pure-input-1' type='text' id='%s' value='%s' name='%s'>\n", a.ID, a.Value, a.Name)
	case "radio":
		fmt.Fprintf(out, "<label for='%s' class='pure-radio'>", a.ID)
		if a.Value != "" {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='radio' id='%s' value='%s' name='%s' checked>\n", a.ID, a.ID, a.Name)
		} else {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='radio' id='%s' value='%s' name='%s'>\n", a.ID, a.ID, a.Name)
		}
		fmt.Fprintf(out, "<span>%s</span>\n", a.Description)
		out.WriteString("</label>")
	default:
		out.WriteString("ERROR: UNKNOWN TYPE " + a.Type)
	}

	if a.Value != "" && a.HasFilledSubattributes() {
		out.WriteString("<div class='sub-form'>\n")
	} else {
		out.WriteString("<div class='sub-form disabled'>\n")
	}
	for _, attr := range a.Attributes {
		out.WriteString(attr.Form())
	}
	out.WriteString("</div>")
	out.WriteString("</div>")
	return out.String()
}

// ConsumerForm is like Form(), but it does not print attributes that have
// "ManualEnforcment" set to true.
func (a *Attribute) ConsumerForm(m *Metadata) string {

	out := new(strings.Builder)
	out.WriteString("<div>")

	var foundAttr *Attribute
	var found bool

	foundAttr, found = m.GetAttribute(a.ID)

	if found {
		a.Value = foundAttr.Value
	}

	if a.DelegatedEnforcement {
		goto subattributes
	}

	switch a.Type {
	case "checkbox":
		fmt.Fprintf(out, "<label for='%s' class='aligned-checkbox'>", a.ID)
		// If there is a value, we check the checkbox
		if a.Value != "" {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='checkbox' id='%s' value='%s' name='%s' checked>\n", a.ID, "checked", a.Name)
		} else {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='checkbox' id='%s' value='%s' name='%s'>\n", a.ID, "checked", a.Name)
		}
		fmt.Fprintf(out, "<span>%s</span>\n", a.Description)
		out.WriteString("</label>")
	case "text":
		fmt.Fprintf(out, "<label>%s</label>\n", a.Description)
		fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" class='pure-input-1' type='text' id='%s' value='%s' name='%s'>\n", a.ID, a.Value, a.Name)
	case "radio":
		fmt.Fprintf(out, "<label for='%s' class='pure-radio'>", a.ID)
		if a.Value != "" {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='radio' id='%s' value='%s' name='%s' checked>\n", a.ID, a.ID, a.Name)
		} else {
			fmt.Fprintf(out, "<input onchange=\"toggleSubform(this)\" type='radio' id='%s' value='%s' name='%s'>\n", a.ID, a.ID, a.Name)
		}
		fmt.Fprintf(out, "<span>%s</span>\n", a.Description)
		out.WriteString("</label>")
	default:
		out.WriteString("ERROR: UNKNOWN TYPE " + a.Type)
	}

	if found && foundAttr.HasFailedReasons() {
		out.WriteString("<div class='failed-reasons'>\n")
		for _, fr := range foundAttr.FailedReasons.FailedReasons {
			out.WriteString("<div class='failed-reason'>\n")
			fmt.Fprintf(out, "<p>Validation failed for dataset %s...:<br>\n", fr.Dataset[:5])
			fmt.Fprintf(out, "<span class='reason'>%s</span></p>\n", fr.Reason)
			out.WriteString("</div>\n")
		}
		out.WriteString("</div>\n")
	}

subattributes:

	subAttrStr := new(strings.Builder)
	for _, attr := range a.Attributes {
		subAttrStr.WriteString(attr.ConsumerForm(m))
	}

	// We must do this after making the recursive call because the value field
	// of the attributes are filled "on the fly". Making the call of
	// HasFilledSubattributes would then always return false because the value
	// fields are not set yet.
	if a.Value != "" && a.HasFilledSubattributes() {
		out.WriteString("<div class='sub-form'>\n")
	} else {
		out.WriteString("<div class='sub-form disabled'>\n")
	}

	out.WriteString(subAttrStr.String())

	out.WriteString("</div>")
	out.WriteString("</div>")
	return out.String()
}

// Darc outputs a DARC string representation of this attribute. Text field must
// have uniq ids acroos different datasets, that is why we append the
// calypsoWriteID. Does NOT browse its sub-attributes
func (a Attribute) Darc(calypsoWriteID string) string {
	out := new(strings.Builder)
	if a.Value != "" && (a.Type == "checkbox" || a.Type == "radio") {
		fmt.Fprintf(out, a.ID+"="+a.Value+"&")
	} else if a.Value != "" && a.Type == "text" {
		fmt.Fprint(out, a.ID+"_"+calypsoWriteID+"="+url.QueryEscape(a.Value)+"&")
	}
	return out.String()
}

// HasFilledSubattributes checks if any of the sub-aatributes of this attribute
// has a value. This is convenient to gray out a sub form in case there isn't
// any sub-attributes filled.
func (a Attribute) HasFilledSubattributes() bool {
	for _, attr := range a.Attributes {
		if attr.Value != "" || attr.HasFilledSubattributes() {
			return true
		}
	}
	return false
}

// AddFailedReason safely adds a reason to the list of failed reasons
func (a *Attribute) AddFailedReason(attributeID, reason, dataset string) {
	if a.FailedReasons == nil {
		a.FailedReasons = &FailedReasons{}
	}
	if a.FailedReasons.FailedReasons == nil {
		a.FailedReasons.FailedReasons = make([]*FailedReason, 0)
	}
	a.FailedReasons.AddReason(attributeID, reason, dataset)
}

// ResetFailedReasons set the FailedReason field of this attribute and its
// subattributes to an empty slice
func (a *Attribute) ResetFailedReasons() {
	if a.FailedReasons == nil {
		a.FailedReasons = &FailedReasons{FailedReasons: []*FailedReason{}}
	}
	for _, attr := range a.Attributes {
		attr.ResetFailedReasons()
	}
}

// HasFailedReasons returns true if it has any failed reasons
func (a Attribute) HasFailedReasons() bool {
	return a.FailedReasons != nil && !a.FailedReasons.IsEmpty()
}

// DelegatedEnforcement describes attributes that validates other attributes
// (normally the ones that have the "delegated_enforcement" attribute to true)
type DelegatedEnforcement struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Attributes  []*EnforcementAttribute `json:"attributes"`
}

func (m DelegatedEnforcement) String() string {
	out := new(strings.Builder)
	out.WriteString("- DelegatedEnforcement:\n")
	fmt.Fprintf(out, "-- Description: %s\n", m.Description)
	out.WriteString("-- Attributes:\n")
	for i, attr := range m.Attributes {
		fmt.Fprintf(out, "--- Attributes[%d]:\n", i)
		out.WriteString(eachLine.ReplaceAllString(attr.String(), "---$1"))
	}

	return out.String()
}

// EnforcementAttribute is a particular type of attribute that valides other
// attributes. It is only a check box that, one checked, whill validates the
// attributes contained in the "Validates" field.
type EnforcementAttribute struct {
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	ValueFromID    string   `json:"value_from_id"`
	TriggerID      string   `json:"trigger_id"`
	TriggerValue   string   `json:"trigger_value"`
	CheckValidates []string `json:"check_validates"`
	TextValidates  string   `json:"text_validates"`
}

func (e EnforcementAttribute) String() string {
	out := new(strings.Builder)
	out.WriteString("- EnforcementAttribute:\n")
	fmt.Fprintf(out, "-- ID: %s\n", e.ID)
	fmt.Fprintf(out, "-- Description: %s\n", e.Description)
	fmt.Fprintf(out, "-- ValueFromID: %s\n", e.ValueFromID)
	fmt.Fprintf(out, "-- TriggerID: %s\n", e.TriggerID)
	fmt.Fprintf(out, "-- TriggerValue: %s\n", e.TriggerValue)
	fmt.Fprintf(out, "-- CheckValidates: %v\n", e.CheckValidates)
	fmt.Fprintf(out, "-- TextValidates: %v\n", e.TextValidates)

	return out.String()
}

// FailedReasons is a simple struct that holds a list a failed reasons
type FailedReasons struct {
	FailedReasons []*FailedReason `json:"failed_reasons"`
}

// AddReason adds a reason to the list of failed reasons
func (f *FailedReasons) AddReason(attributeID, reason, dataset string) {
	if f.FailedReasons == nil {
		f.FailedReasons = make([]*FailedReason, 0)
	}
	f.FailedReasons = append(f.FailedReasons, &FailedReason{
		AttributeID: attributeID,
		Reason:      reason,
		Dataset:     dataset,
	})
}

// IsEmpty returns true if there is
func (f FailedReasons) IsEmpty() bool {
	return f.FailedReasons == nil || len(f.FailedReasons) == 0
}

func (f FailedReasons) String() string {
	out := new(strings.Builder)
	out.WriteString("- FailedReasons:\n")
	for i, fr := range f.FailedReasons {
		fmt.Fprintf(out, "-- FailedReasons[%d]\n", i)
		out.WriteString(eachLine.ReplaceAllString(fr.String(), "--$1"))
	}
	return out.String()
}

// FailedReason describes a failed reason for a specific attribute
type FailedReason struct {
	AttributeID string `json:"attribute_id"`
	Reason      string `json:"reason"`
	Dataset     string `json:"dataset"`
}

// String returns the string representation of a failed reason
func (f FailedReason) String() string {
	out := new(strings.Builder)
	out.WriteString("- FailedReason:\n")
	fmt.Fprintf(out, "-- AttributeID: %s\n", f.AttributeID)
	fmt.Fprintf(out, "-- Reason: %s\n", f.Reason)
	fmt.Fprintf(out, "-- Dataset: %s\n", f.Dataset)
	return out.String()
}

// AuditData is used by the catadmin cli to return audit infos
type AuditData struct {
	// BlocksChecked is the number of blocks checked
	BlocksChecked int
	// OccFound is the number of blocks concerned
	OccFound int
	// Blocks contains the list of audit blocks
	Blocks []*AuditBlock
}

// AuditBlock is a tailored version of a skipblock that we need to display
// something interesting
type AuditBlock struct {
	BlockIndex int
	// The number of blocks before the previous in the list. If the current block
	// has an index of 7 and the previous one in the list has an index of 5, then
	// the delta is 1. This is handy for go templates where we cannot do
	// aritmetic operations. The first one has a value of -1.
	DeltaPrevious int
	// The number of blocks after the next in the list. The last one has a value
	// of -1.
	DeltaNext    int
	Transactions []*AuditTransaction
}

// AuditTransaction is a tailored version of a transaction for audit
type AuditTransaction struct {
	Accepted     bool
	Instructions []*byzcoin.Instruction
}
