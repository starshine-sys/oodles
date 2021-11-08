package db

// ArikawaDocumentationBase ...
const ArikawaDocumentationBase = "https://pkg.go.dev/github.com/diamondburned/arikawa/v3"

// ConfigOptions are all configuration options
var ConfigOptions = map[string]ConfigOption{
	// core bot configuration
	"prefix": {
		Description:  "The prefix used for bot commands. Case-insensitive.",
		Type:         StringOptionType,
		DefaultValue: ".",
	},
	"activity": {
		Description:  "The activity shown in the bot's status.",
		Type:         StringOptionType,
		DefaultValue: "",
	},
	"activity_type": {
		Description:  "The activity type shown in the bot's status. Valid options are: `playing`, `listening`, `watching`",
		Type:         StringOptionType,
		DefaultValue: "playing",
		ValidValues:  []interface{}{"playing", "listening", "watching"},
	},
	"status": {
		Description:  "The bot's status. Valid options are: `online`, `idle`, `dnd`",
		Type:         StringOptionType,
		DefaultValue: "online",
		ValidValues:  []interface{}{"online", "idle", "dnd"},
	},

	// welcome configuration
	"welcome_channel": {
		Description:  "The channel new joins are announced in, and where they are welcomed if approved.",
		Type:         SnowflakeOptionType,
		DefaultValue: 0,
	},
	"welcome_message": {
		Description: "The message used to welcome someone. Accepted variables are:\n- `{{.Guild}}`: the [guild]({docs}/discord#Guild) the user is welcomed in\n- `{{.Member}}`: the [member]({docs}/discord#Member) that is being welcomed\n- `{{.Approver}}`: the [member]({docs}/discord#Member) that approved the user\n\nThis message is sent to `welcome_channel`.",

		Type:         StringOptionType,
		DefaultValue: "Welcome to {{.Guild.Name}}, {{.Member.User.Mention}}!",
	},
	"deny_message": {
		Description: "The message sent in `welcome_channel` when a user is denied. If empty, denials will not be posted publicly.\nAccepted variables are:\n- `{{.Guild}}`: the [guild]({docs}/discord#Guild) the user is denied in\n- `{{.User}}`: the [user]({docs}/discord#User) that was denied\n- `{{.Denier}}`: the [member]({docs}/discord#Member) that denied the user\n- `{{.Reason}}`: the reason the user was denied, or \"No reason specified\" if no reason was given.",

		Type:         StringOptionType,
		DefaultValue: "{{.User.Mention}} ({{.User.Tag}}) was denied entry to the server by {{.Denier.User.Username}}.\nReason: {{.Reason}}",
	},

	// application configuration
	"application_category": {
		Description:  "The category where application channels are created.",
		Type:         SnowflakeOptionType,
		DefaultValue: 0,
	},
	"discussion_channel": {
		Description:  "The channel where newly finished applications are announced.",
		Type:         SnowflakeOptionType,
		DefaultValue: 0,
	},
	"application_channel_message": {
		Description:  "The message that is posted in the applications channel. `{guild}` is replaced with the guild name.",
		Type:         StringOptionType,
		DefaultValue: "Thank you for joining {guild}!\nWe hope you enjoy your stay.",
	},
	"open_application_message": {
		Description:  "The message that is posted when an application is opened. `{guild}` is replaced with the guild name.",
		Type:         StringOptionType,
		DefaultValue: "Thank you for beginning your application for {guild}.",
	},
	"application_finished_message": {
		Description:  "The message that is posted when an application finishes.",
		Type:         StringOptionType,
		DefaultValue: "Application finished! Please continue to add proof if it included more than one screenshot, and otherwise, please be patient until a mod can review your answers.",
	},
	"long_answer_minimum": {
		Description:  "Minimum number of words needed to advance questions where long answers are required.",
		Type:         IntOptionType,
		DefaultValue: 3,
	},
	"long_answer_message": {
		Description:  "Message sent when someone's answer is too short. `{num}` is replaced with the number of words required.",
		Type:         StringOptionType,
		DefaultValue: "Sorry, but your answer is too short! Please resend it, and make sure it's at least {num} words long.",
	},

	// verification configuration
	"verified_role": {
		Description:  "The role given to a member when they are approved.",
		Type:         SnowflakeOptionType,
		DefaultValue: 0,
	},
	"adult_role": {
		Description:  "The role given to an adult member when they are approved.\n\nIf either this or `minor_role` is invalid, this step will be skipped and the valid role will be given.\nIf both this and `minor_role` are invalid, only `verified_role` will be given when a member is approved.",
		Type:         SnowflakeOptionType,
		DefaultValue: 0,
	},
	"minor_role": {
		Description:  "The role given to a minor member when they are approved.\n\nIf either this or `adult_role` is invalid, this step will be skipped and the valid role will be given.\nIf both this and `adult_role` are invalid, only `verified_role` will be given when a member is approved.",
		Type:         SnowflakeOptionType,
		DefaultValue: 0,
	},
	"keep_application_visible": {
		Description:  "Whether to keep the application channel visible to a member once they are approved.\n(If they are denied, the channel is hidden immediately)",
		Type:         BoolOptionType,
		DefaultValue: false,
	},
	"kick_on_deny": {
		Description:  "Whether to kick a user immediately after they are denied with the `{prefix}deny` command. See also: `confirm_deny`, `dm_on_deny`.",
		Type:         BoolOptionType,
		DefaultValue: false,
	},
	"dm_on_deny": {
		Description:  "Whether to DM a user if they are denied with the `{prefix}deny` command. See also: `confirm_deny`, `kick_on_deny`.",
		Type:         BoolOptionType,
		DefaultValue: false,
	},
	"confirm_deny": {
		Description:  "Whether to show a confirmation prompt for the `{prefix}deny` command. See also: `dm_on_deny`, `kick_on_deny`.",
		Type:         BoolOptionType,
		DefaultValue: false,
	},
}
