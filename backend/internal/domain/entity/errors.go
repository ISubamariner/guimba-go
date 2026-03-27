package entity

import "errors"

// Domain errors for Program entity.
var (
	ErrProgramNameRequired   = errors.New("program name is required")
	ErrProgramNameTooLong    = errors.New("program name must be 255 characters or less")
	ErrProgramInvalidStatus  = errors.New("program status is invalid")
	ErrProgramEndBeforeStart = errors.New("end date cannot be before start date")
)

// Domain errors for User entity.
var (
	ErrUserEmailRequired      = errors.New("user email is required")
	ErrUserEmailTooLong       = errors.New("user email must be 255 characters or less")
	ErrUserFullNameRequired   = errors.New("user full name is required")
	ErrUserFullNameTooLong    = errors.New("user full name must be 255 characters or less")
	ErrUserPasswordRequired   = errors.New("user password is required")
	ErrUserNotActive          = errors.New("user account is not active")
	ErrUserInvalidCredentials = errors.New("invalid email or password")
)

// Domain errors for Beneficiary entity.
var (
	ErrBeneficiaryFullNameRequired = errors.New("beneficiary full name is required")
	ErrBeneficiaryFullNameTooLong  = errors.New("beneficiary full name must be 255 characters or less")
	ErrBeneficiaryInvalidStatus    = errors.New("beneficiary status is invalid")
	ErrBeneficiaryContactRequired  = errors.New("beneficiary must have at least one contact method (email or phone)")
	ErrBeneficiaryAlreadyEnrolled  = errors.New("beneficiary is already enrolled in this program")
	ErrBeneficiaryNotEnrolled      = errors.New("beneficiary is not enrolled in this program")
)

// Domain errors for Role entity.
var (
	ErrRoleNameRequired        = errors.New("role name is required")
	ErrRoleNameTooLong         = errors.New("role name must be 50 characters or less")
	ErrRoleDisplayNameRequired = errors.New("role display name is required")
	ErrRoleSystemCannotDelete  = errors.New("system roles cannot be deleted")
)

// Domain errors for Tenant entity.
var (
	ErrTenantFullNameRequired = errors.New("tenant full name is required")
	ErrTenantFullNameTooLong  = errors.New("tenant full name must be 255 characters or less")
	ErrTenantContactRequired  = errors.New("tenant must have at least one contact method (email or phone)")
	ErrTenantEmailExists      = errors.New("a tenant with this email already exists")
)

// Domain errors for Property entity.
var (
	ErrPropertyNameRequired = errors.New("property name is required")
	ErrPropertyNameTooLong  = errors.New("property name must be 255 characters or less")
	ErrPropertyCodeRequired = errors.New("property code is required")
	ErrPropertySizeRequired = errors.New("property size in square meters must be greater than zero")
	ErrPropertyCodeExists   = errors.New("a property with this code already exists")
)

// Domain errors for Money value object.
var (
	ErrCurrencyMismatch   = errors.New("currency mismatch: operations require same currency")
	ErrNegativeAmount     = errors.New("amount must not be negative")
	ErrInvalidCurrency    = errors.New("invalid currency code")
	ErrInsufficientAmount = errors.New("insufficient amount for operation")
)

// Domain errors for Debt entity.
var (
	ErrDebtDescriptionRequired      = errors.New("debt description is required")
	ErrDebtDescriptionTooLong       = errors.New("debt description must be 500 characters or less")
	ErrDebtAmountRequired           = errors.New("debt amount must be greater than zero")
	ErrDebtInvalidType              = errors.New("invalid debt type")
	ErrDebtDueDateRequired          = errors.New("debt due date is required")
	ErrDebtAlreadyPaid              = errors.New("debt is already fully paid")
	ErrDebtAlreadyCancelled         = errors.New("debt is already cancelled")
	ErrDebtOverpayment              = errors.New("payment amount exceeds remaining balance")
	ErrDebtInvalidStateTransition   = errors.New("invalid debt state transition")
	ErrDebtNotPayable               = errors.New("debt is not in a payable state")
	ErrPropertyHasActiveDebts       = errors.New("cannot deactivate property with active debts")
)

// Domain errors for enum validation.
var (
	ErrInvalidDebtStatus        = errors.New("invalid debt status")
	ErrInvalidDebtType          = errors.New("invalid debt type value")
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
	ErrInvalidPaymentMethod     = errors.New("invalid payment method")
)

// Domain errors for Transaction entity.
var (
	ErrTransactionAmountRequired       = errors.New("transaction amount must be greater than zero")
	ErrTransactionInvalidType          = errors.New("invalid transaction type")
	ErrTransactionInvalidPaymentMethod = errors.New("invalid payment method")
	ErrTransactionDateRequired         = errors.New("transaction date is required")
	ErrTransactionAlreadyVerified      = errors.New("transaction is already verified")
	ErrTransactionDuplicateReference   = errors.New("duplicate reference number for this debt")
)
