export class Utils {

    /**
     * Return a number in the range 0..22 based on the given string's content.
     * @param s String to calculate colour index for.
     */
    static colourIndex(s: string) {
        return [...s].reduce((sum, c) => sum + c.charCodeAt(0), 0) % 23;
    }

    /**
     * Return a string representation of a time difference in the "time ago" notation.
     * @param current Current time in milliseconds.
     * @param previous The past moment in milliseconds.
     */
    static timeAgo(current: number, previous: number): string {
        const seconds = Math.floor((current-previous) / 1000);

        // Years
        let interval = Math.floor(seconds / 31536000);
        if (interval > 1) {
            return `${interval} years ago`;
        }
        if (interval === 1) {
            return 'A year ago';
        }

        // Months
        interval = Math.floor(seconds / 2592000);
        if (interval > 1) {
            return `${interval} months ago`;
        }
        if (interval === 1) {
            return 'A month ago';
        }

        // Days
        interval = Math.floor(seconds / 86400);
        if (interval > 1) {
            return `${interval} days ago`;
        }
        if (interval === 1) {
            return 'Yesterday';
        }

        // Hours
        interval = Math.floor(seconds / 3600);
        if (interval > 1) {
            return `${interval} hours ago`;
        }
        if (interval === 1) {
            return 'An hour ago';
        }

        // Minutes
        interval = Math.floor(seconds / 60);
        if (interval > 1) {
            return `${interval} minutes ago`;
        }
        if (interval === 1) {
            return 'A minute ago';
        }

        // Less than a minute
        return 'Just now';
    }
}
