export type StringBooleanMap = { [k: string]: boolean };

export interface Comment {
    readonly commentHex:   string;
    readonly commenterHex: string;
    readonly parentHex:    string;
    readonly creationDate: string;

    // Mutable
    state:     'approved' | 'unapproved' | 'flagged';
    deleted:   boolean;
    direction: number;
    score:     number;
    markdown?: string;
    html?:     string;

    // Computed
    creationMs?: number;
}

export interface Commenter {
    readonly commenterHex?: string;
    readonly email?:        string;
    readonly name?:         string;
    readonly websiteUrl?:   string;
    readonly avatarUrl?:    string;
    readonly provider?:     string;
    readonly joinDate?:     string;

    // Mutable
    isModerator?: boolean;
}

export interface Email {
    readonly email?:                      string;
    readonly unsubscribeSecretHex?:       string;
    readonly lastEmailNotificationDate?:  string;

    // Mutable
    sendReplyNotifications?:     boolean;
    sendModeratorNotifications?: boolean;
}

export type CommentsGroupedByHex = { [k: string]: Comment[] };

export type CommenterMap = { [k: string]: Commenter };

export type ComparatorFunc<T> = (a: T, b: T) => number;

export type SortPolicy = 'score-desc' | 'creationdate-desc' | 'creationdate-asc';

export interface SortPolicyProps<T> {
    label:      string;
    comparator: ComparatorFunc<T>;
}

export interface SignupData {
    email:       string;
    name:        string;
    password:    string;
    websiteUrl?: string;
}

export interface ProfileSettings {
    email:           string;  // Commenter's email address
    name:            string;  // Commenter's name
    websiteUrl?:     string;  // Optional commenter's website URL
    avatarUrl?:      string;  // Optional commenter's avatar URL
    notifyModerator: boolean; // Whether to send moderator notifications to the user
    notifyReplies:   boolean; // Whether to send reply notifications to the user
}

export const AnonymousCommenterId = '0000000000000000000000000000000000000000000000000000000000000000';

export const sortingProps: { [k in SortPolicy]: SortPolicyProps<Comment> } = {
    'score-desc':        {label: 'Upvotes', comparator: (a, b) => b.score - a.score},
    'creationdate-desc': {label: 'Newest',  comparator: (a, b) => a.creationMs! < b.creationMs! ? 1 : -1},
    'creationdate-asc':  {label: 'Oldest',  comparator: (a, b) => a.creationMs! < b.creationMs! ? -1 : 1},
};

