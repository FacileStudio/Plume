package mcp

// listInput is plume_list_documents's input.
type listInput struct {
	SpaceID  int64  `json:"space_id,omitempty" jsonschema:"a space id from spaces in a previous result; default is your personal documents"`
	Status   string `json:"status,omitempty" jsonschema:"only documents in this state: draft, pending, completed or declined"`
	ClientID int64  `json:"client_id,omitempty" jsonschema:"only documents filed under this client, from plume_list_clients"`
	Query    string `json:"query,omitempty" jsonschema:"keep documents whose name contains this, case ignored"`
	Cursor   string `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page"`
}

// documentInput names one document.
type documentInput struct {
	DocumentID int64 `json:"document_id" jsonschema:"a document id from plume_list_documents"`
}

// signerInput is one person who signs, approves or only receives a copy.
type signerInput struct {
	Name  string `json:"name" jsonschema:"the person's full name"`
	Email string `json:"email" jsonschema:"where Plume emails the signing link"`
	Role  string `json:"role,omitempty" jsonschema:"signer (default) signs, approver approves, viewer only receives the completed document"`
	Order int    `json:"order,omitempty" jsonschema:"position when the document is sequential: lower signs first. Default 0"`
}

// createInput is plume_create_document's input.
type createInput struct {
	Name       string        `json:"name" jsonschema:"the document's title, shown to the signers"`
	FileName   string        `json:"file_name,omitempty" jsonschema:"the PDF's file name, such as contract.pdf. Default is the name plus .pdf"`
	PDFBase64  string        `json:"pdf_base64" jsonschema:"the PDF file, base64 encoded, at most 3 MB once decoded"`
	SpaceID    int64         `json:"space_id,omitempty" jsonschema:"file it in this shared space; default is your personal documents"`
	ClientID   int64         `json:"client_id,omitempty" jsonschema:"file it under this client, from plume_list_clients"`
	Sequential bool          `json:"sequential,omitempty" jsonschema:"signers sign one after another by order instead of all at once"`
	Signers    []signerInput `json:"signers" jsonschema:"1 to 20 people; each signer and approver gets a signature box on the last page"`
}

// updateInput is plume_update_document's input: only the fields given change.
type updateInput struct {
	DocumentID int64  `json:"document_id" jsonschema:"a draft document id"`
	Name       string `json:"name,omitempty" jsonschema:"the new title"`
	Sequential *bool  `json:"sequential,omitempty" jsonschema:"true makes signers sign one after another by order"`
	ClientID   *int64 `json:"client_id,omitempty" jsonschema:"file it under this client; 0 removes the client"`
}

// addSignerInput is plume_add_signer's input.
type addSignerInput struct {
	DocumentID int64 `json:"document_id" jsonschema:"a draft document id"`
	signerInput
	NoSignatureBox bool `json:"no_signature_box,omitempty" jsonschema:"do not place a signature box for this person; they then have nothing to fill in"`
}

// signerRef names one signer.
type signerRef struct {
	SignerID int64 `json:"signer_id" jsonschema:"a signer id from plume_get_document"`
}

// clientRef names one client.
type clientRef struct {
	ClientID int64 `json:"client_id" jsonschema:"a client id from plume_list_clients"`
}

// fieldRef names one field.
type fieldRef struct {
	FieldID int64 `json:"field_id" jsonschema:"a field id from the fields of plume_get_document"`
}

// clientsInput is plume_list_clients's input.
type clientsInput struct {
	Query  string `json:"query,omitempty" jsonschema:"keep clients whose name, company or email contains this, case ignored"`
	Cursor string `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page"`
}

// saveClientInput is plume_save_client's input: a new client without client_id,
// otherwise only the fields given change.
type saveClientInput struct {
	ClientID int64   `json:"client_id,omitempty" jsonschema:"the client to change; leave out to create one"`
	Name     string  `json:"name,omitempty" jsonschema:"the client's name, required for a new client"`
	Email    *string `json:"email,omitempty" jsonschema:"the client's email; an empty string clears it"`
	Company  *string `json:"company,omitempty" jsonschema:"the client's company; an empty string clears it"`
	Phone    *string `json:"phone,omitempty" jsonschema:"the client's phone; an empty string clears it"`
	Notes    *string `json:"notes,omitempty" jsonschema:"free notes about the client; an empty string clears them"`
}
