package bo

// UserBO represents the business object for a user.
// It encapsulates user-related information used across different layers of the application.
type UserBO struct {
	// Id is the unique identifier for the user.
	Id int64
	// Num is a user-specific number, which might be used for various purposes like internal tracking.
	Num string
	// Account is the user's login account name.
	Account string
	// Password is the user's password, typically stored in a hashed format.
	Password string
	// Nickname is the user's display name.
	Nickname string
	// Avatar is the URL pointing to the user's profile picture.
	Avatar string
	// RegisterSource indicates where the user registered from, e.g., website, mobile app.
	RegisterSource int32
	// Stats represents the user's status, such as active, inactive, or banned.
	Stats int32
	// Token is the authentication token used to identify the user in subsequent requests.
	Token string
}
