export interface Comment {
    readonly commentHex?:   string;
    readonly commenterHex?: string;
    readonly parentHex?:    string;
    readonly score?:        number;
    readonly state?:        string;
    readonly creationDate?: string;
    readonly direction?:    number;
    readonly deleted?:      boolean;

    // Mutable
    markdown?: string;
    html?:     string;

    // Computed
    creationMs?: number;
}

export interface Commenter {
    readonly commenterHex?: string;
    readonly name?:         string;
    readonly link?:         string;
    readonly photo?:        string;
    readonly provider?:     string;
    readonly joinDate?:     string;
    readonly isModerator?:  boolean;
}

export interface Email {
    readonly email?:                      string;
    readonly unsubscribeSecretHex?:       string;
    readonly lastEmailNotificationDate?:  string;
    readonly sendReplyNotifications?:     boolean;
    readonly sendModeratorNotifications?: boolean;
}

export type CommentMap = { [k: string]: Comment };
export type CommentsMap = { [k: string]: Comment[] };

export interface OAuthResponse {
    provider?: string;
    id?:       string;
}

export type SortPolicy = 'score-desc' | 'creationdate-desc' | 'creationdate-asc';
