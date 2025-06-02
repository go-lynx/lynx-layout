package bo

// UserBO represents the business object for a user.
// UserBO 表示用户的业务对象。
// It encapsulates user-related information used across different layers of the application.
// 它封装了在应用程序不同层使用的用户相关信息。
type UserBO struct {
	// Id is the unique identifier for the user.
	// Id 是用户的唯一标识符。
	Id int64
	// Num is a user-specific number, which might be used for various purposes like internal tracking.
	// Num 是用户特定的编号，可能用于内部跟踪等各种用途。
	Num string
	// Account is the user's login account name.
	// Account 是用户的登录账号名。
	Account string
	// Password is the user's password, typically stored in a hashed format.
	// Password 是用户的密码，通常以哈希格式存储。
	Password string
	// Nickname is the user's display name.
	// Nickname 是用户的显示昵称。
	Nickname string
	// Avatar is the URL pointing to the user's profile picture.
	// Avatar 是指向用户头像图片的 URL。
	Avatar string
	// RegisterSource indicates where the user registered from, e.g., website, mobile app.
	// RegisterSource 指示用户的注册来源，例如网站、移动应用等。
	RegisterSource int32
	// Stats represents the user's status, such as active, inactive, or banned.
	// Stats 表示用户的状态，如活跃、非活跃或封禁。
	Stats int32
	// Token is the authentication token used to identify the user in subsequent requests.
	// Token 是用于在后续请求中识别用户的认证令牌。
	Token string
}
