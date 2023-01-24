export type StringBooleanMap = { [k: string]: boolean };

export interface Comment {
    readonly commentHex:   string;
    readonly commenterHex: string;
    readonly parentHex:    string;
    readonly state:        string;
    readonly creationDate: string;
    readonly direction:    number;
    readonly deleted:      boolean;

    // Mutable
    score:     number;
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
export type CommentsGroupedByHex = { [k: string]: Comment[] };

export type ComparatorFunc<T> = (a: T, b: T) => number;

export type SortPolicy = 'score-desc' | 'creationdate-desc' | 'creationdate-asc';

export interface SortPolicyProps<T> {
    label:      string;
    comparator: ComparatorFunc<T>;
}
