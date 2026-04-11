package dto

type DiagnosisUpdateRequest struct {
	Code        *string `json:"code"        validate:"omitempty,code,max=50"`
	Description *string `json:"description" validate:"omitempty,min=5,max=4000"`
}
