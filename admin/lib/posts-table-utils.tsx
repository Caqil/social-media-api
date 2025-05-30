// lib/posts-table-utils.tsx - Enhanced table utilities for posts
import React from "react";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  IconPhoto,
  IconVideo,
  IconMusic,
  IconFile,
  IconGlobe,
  IconUsers,
  IconLock,
  IconHeart,
  IconMessage,
  IconShare,
  IconEye,
  IconFlag,
  IconCalendar,
  IconMapPin,
  IconHash,
  IconAt,
  IconActivity,
  IconTrendingUp,
  IconClock,
  IconLink,
  IconEyeOff,
} from "@tabler/icons-react";
import { Post, PrivacyLevel, UserResponse } from "@/types/admin";
import { PlayIcon } from "lucide-react";

// Post-specific formatters
export const PostTableFormatters = {
  // Post author with profile
  postAuthor: (user: UserResponse, showVerified = true) => (
    <div className="flex items-center gap-3">
      <Avatar className="h-8 w-8">
        <AvatarImage
          src={user?.profile_picture}
          alt={user?.username || "User"}
        />
        <AvatarFallback>
          {(user?.first_name?.[0] || user?.username?.[0] || "U").toUpperCase()}
        </AvatarFallback>
      </Avatar>
      <div className="min-w-0">
        <div className="font-medium flex items-center gap-2">
          {user?.first_name
            ? `${user.first_name} ${user.last_name || ""}`.trim()
            : user?.username || "Unknown User"}
          {showVerified && user?.is_verified && (
            <div className="w-3 h-3 bg-blue-500 rounded-full flex items-center justify-center">
              <div className="w-1.5 h-1.5 bg-white rounded-full" />
            </div>
          )}
        </div>
        <div className="text-sm text-muted-foreground">
          @{user?.username || "unknown"}
        </div>
      </div>
    </div>
  ),

  // Post content preview with media indicator
  postContent: (post: Post, maxLength = 100) => (
    <div className="max-w-md space-y-2">
      {/* Post type indicator */}
      <div className="flex items-center gap-2">
        {PostTableFormatters.getPostTypeIcon(post.type)}
        <span className="text-xs text-muted-foreground capitalize">
          {post.type || "text"} post
        </span>
        {post.is_pinned && (
          <Badge variant="outline" className="text-xs px-1 py-0">
            Pinned
          </Badge>
        )}
      </div>

      {/* Content preview */}
      <div className="text-sm">
        {post.content ? (
          <span className="line-clamp-2">
            {post.content.length > maxLength
              ? `${post.content.substring(0, maxLength)}...`
              : post.content}
          </span>
        ) : (
          <span className="text-muted-foreground italic">No text content</span>
        )}
      </div>

      {/* Media indicators */}
      {post.media_urls && post.media_urls.length > 0 && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <IconPhoto className="h-3 w-3" />
          <span>
            +{post.media_urls.length} media file
            {post.media_urls.length > 1 ? "s" : ""}
          </span>
        </div>
      )}

      {/* Hashtags preview */}
      {post.hashtags && post.hashtags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {post.hashtags.slice(0, 3).map((tag, index) => (
            <Badge key={index} variant="secondary" className="text-xs">
              #{tag}
            </Badge>
          ))}
          {post.hashtags.length > 3 && (
            <span className="text-xs text-muted-foreground">
              +{post.hashtags.length - 3} more
            </span>
          )}
        </div>
      )}

      {/* Location indicator */}
      {post.location && (
        <div className="flex items-center gap-1 text-xs text-muted-foreground">
          <IconMapPin className="h-3 w-3" />
          <span className="truncate">{post.location}</span>
        </div>
      )}
    </div>
  ),

  // Post type icon
  getPostTypeIcon: (type: string) => {
    const iconProps = { className: "h-4 w-4" };

    switch (type?.toLowerCase()) {
      case "image":
        return <IconPhoto {...iconProps} className="h-4 w-4 text-green-600" />;
      case "video":
        return <IconVideo {...iconProps} className="h-4 w-4 text-blue-600" />;
      case "audio":
        return <IconMusic {...iconProps} className="h-4 w-4 text-purple-600" />;
      case "poll":
        return (
          <IconActivity {...iconProps} className="h-4 w-4 text-orange-600" />
        );
      case "shared":
        return <IconShare {...iconProps} className="h-4 w-4 text-indigo-600" />;
      case "live":
        return <PlayIcon {...iconProps} className="h-4 w-4 text-red-600" />;
      default:
        return <IconMessage {...iconProps} className="h-4 w-4 text-gray-600" />;
    }
  },

  // Post visibility with icon and description
  postVisibility: (visibility: PrivacyLevel) => {
    const getVisibilityConfig = (level: PrivacyLevel) => {
      switch (level) {
        case PrivacyLevel.PUBLIC:
          return {
            icon: IconGlobe,
            color: "bg-green-100 text-green-800 border-green-200",
            description: "Everyone can see",
          };
        case PrivacyLevel.FRIENDS:
          return {
            icon: IconUsers,
            color: "bg-blue-100 text-blue-800 border-blue-200",
            description: "Friends only",
          };
        case PrivacyLevel.PRIVATE:
          return {
            icon: IconLock,
            color: "bg-gray-100 text-gray-800 border-gray-200",
            description: "Private",
          };
        default:
          return {
            icon: IconGlobe,
            color: "bg-green-100 text-green-800 border-green-200",
            description: "Public",
          };
      }
    };

    const config = getVisibilityConfig(visibility || PrivacyLevel.PUBLIC);
    const Icon = config.icon;

    return (
      <Badge
        variant="outline"
        className={config.color}
        title={config.description}
      >
        <Icon className="h-3 w-3 mr-1" />
        {(visibility || "public").charAt(0).toUpperCase() +
          (visibility || "public").slice(1)}
      </Badge>
    );
  },

  // Enhanced engagement metrics
  postEngagement: (post: Post, showDetails = true) => (
    <div className="space-y-2">
      {/* Main engagement metrics */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1 text-sm">
          <IconHeart className="h-3 w-3 text-red-500" />
          <span>{(post.likes_count || 0).toLocaleString()}</span>
        </div>
        <div className="flex items-center gap-1 text-sm">
          <IconMessage className="h-3 w-3 text-blue-500" />
          <span>{(post.comments_count || 0).toLocaleString()}</span>
        </div>
        <div className="flex items-center gap-1 text-sm">
          <IconShare className="h-3 w-3 text-green-500" />
          <span>{(post.shares_count || 0).toLocaleString()}</span>
        </div>
      </div>

      {/* Views and engagement rate */}
      {showDetails && (
        <div className="space-y-1">
          {post.views_count !== undefined && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <IconEye className="h-3 w-3" />
              <span>{post.views_count.toLocaleString()} views</span>
            </div>
          )}

          {/* Calculate and show engagement rate */}
          {post.views_count && post.views_count > 0 && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <IconTrendingUp className="h-3 w-3" />
              <span>
                {(
                  (((post.likes_count || 0) +
                    (post.comments_count || 0) +
                    (post.shares_count || 0)) /
                    post.views_count) *
                  100
                ).toFixed(1)}
                % engagement
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  ),

  // Post status with comprehensive indicators
  postStatus: (post: Post) => {
    const getStatusElements = () => {
      const elements = [];

      // Primary status
      if (post.is_hidden) {
        elements.push(
          <Badge key="hidden" variant="destructive" className="text-xs">
            Hidden
          </Badge>
        );
      } else {
        elements.push(
          <Badge key="active" variant="default" className="text-xs">
            Active
          </Badge>
        );
      }

      // Secondary indicators
      if (post.is_reported) {
        elements.push(
          <Badge
            key="reported"
            variant="outline"
            className="text-xs text-orange-600 border-orange-200"
          >
            <IconFlag className="h-3 w-3 mr-1" />
            Reported
          </Badge>
        );
      }

      if (post.is_pinned) {
        elements.push(
          <Badge
            key="pinned"
            variant="outline"
            className="text-xs text-blue-600 border-blue-200"
          >
            Pinned
          </Badge>
        );
      }

      if (post.is_promoted) {
        elements.push(
          <Badge
            key="promoted"
            variant="outline"
            className="text-xs text-purple-600 border-purple-200"
          >
            Promoted
          </Badge>
        );
      }

      return elements;
    };

    return <div className="space-y-1">{getStatusElements()}</div>;
  },

  // Post timestamp with relative time
  postTimestamp: (dateString: string, showRelative = true) => {
    if (!dateString) return <span className="text-muted-foreground">-</span>;

    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    let relativeText = "";
    if (showRelative) {
      if (diffMins < 1) relativeText = "Just now";
      else if (diffMins < 60) relativeText = `${diffMins}m ago`;
      else if (diffHours < 24) relativeText = `${diffHours}h ago`;
      else if (diffDays < 7) relativeText = `${diffDays}d ago`;
      else relativeText = date.toLocaleDateString();
    }

    return (
      <div className="flex items-center gap-2 text-sm">
        <IconCalendar className="h-4 w-4 text-muted-foreground" />
        <div>
          <div className="font-medium">
            {showRelative && relativeText
              ? relativeText
              : date.toLocaleDateString()}
          </div>
          <div className="text-xs text-muted-foreground">
            {date.toLocaleTimeString()}
          </div>
        </div>
      </div>
    );
  },

  // Post media preview
  postMediaPreview: (post: Post, maxItems = 3) => {
    if (!post.media_urls || post.media_urls.length === 0) {
      return <span className="text-muted-foreground text-xs">No media</span>;
    }

    return (
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          {PostTableFormatters.getPostTypeIcon(post.type)}
          <span className="text-xs text-muted-foreground">
            {post.media_urls.length} file{post.media_urls.length > 1 ? "s" : ""}
          </span>
        </div>

        {/* Show thumbnails for images */}
        {post.type === "image" && (
          <div className="flex gap-1">
            {post.media_urls.slice(0, maxItems).map((url, index) => (
              <div
                key={index}
                className="w-8 h-8 bg-muted rounded border flex items-center justify-center"
              >
                <IconPhoto className="h-4 w-4 text-muted-foreground" />
              </div>
            ))}
            {post.media_urls.length > maxItems && (
              <div className="w-8 h-8 bg-muted rounded border flex items-center justify-center">
                <span className="text-xs">
                  +{post.media_urls.length - maxItems}
                </span>
              </div>
            )}
          </div>
        )}
      </div>
    );
  },

  // Post mentions formatter
  postMentions: (mentions: string[]) => {
    if (!mentions || mentions.length === 0) {
      return <span className="text-muted-foreground text-xs">No mentions</span>;
    }

    return (
      <div className="flex items-center gap-1">
        <IconAt className="h-3 w-3 text-muted-foreground" />
        <span className="text-xs text-muted-foreground">
          {mentions.length} mention{mentions.length > 1 ? "s" : ""}
        </span>
      </div>
    );
  },

  // Post group/event association
  postAssociation: (post: Post) => {
    if (post.group_id) {
      return (
        <Badge variant="outline" className="text-xs">
          <IconUsers className="h-3 w-3 mr-1" />
          Group Post
        </Badge>
      );
    }

    if (post.event_id) {
      return (
        <Badge variant="outline" className="text-xs">
          <IconCalendar className="h-3 w-3 mr-1" />
          Event Post
        </Badge>
      );
    }

    return null;
  },

  // Post scheduled indicator
  postScheduled: (post: Post) => {
    if (!post.scheduled_at) return null;

    const scheduledDate = new Date(post.scheduled_at);
    const now = new Date();
    const isPast = scheduledDate < now;

    return (
      <Badge
        variant="outline"
        className={`text-xs ${
          isPast
            ? "text-green-600 border-green-200"
            : "text-orange-600 border-orange-200"
        }`}
      >
        <IconClock className="h-3 w-3 mr-1" />
        {isPast ? "Published" : "Scheduled"}
      </Badge>
    );
  },

  // Post edit history indicator
  postEditHistory: (post: Post) => {
    if (!post.edit_history || post.edit_history.length === 0) return null;

    return (
      <Badge
        variant="outline"
        className="text-xs text-gray-600 border-gray-200"
      >
        Edited {post.edit_history.length} time
        {post.edit_history.length > 1 ? "s" : ""}
      </Badge>
    );
  },
};

// Post-specific bulk actions
export const PostBulkActions = [
  {
    label: "Show Posts",
    action: "show",
    icon: IconEye,
    description: "Make selected posts visible to users",
  },
  {
    label: "Hide Posts",
    action: "hide",
    icon: IconEyeOff,
    description: "Hide selected posts from users",
  },
  {
    label: "Pin Posts",
    action: "pin",
    icon: IconTrendingUp,
    description: "Pin selected posts to top",
  },
  {
    label: "Unpin Posts",
    action: "unpin",
    icon: IconTrendingUp,
    description: "Remove pin from selected posts",
  },
  {
    label: "Delete Posts",
    action: "delete",
    icon: IconFlag,
    variant: "destructive" as const,
    description: "Permanently delete selected posts",
  },
];

// Post filter options
export const PostFilterOptions = {
  type: [
    { label: "All Types", value: "all", icon: IconEye },
    { label: "Text", value: "text", icon: IconMessage },
    { label: "Image", value: "image", icon: IconPhoto },
    { label: "Video", value: "video", icon: IconVideo },
    { label: "Audio", value: "audio", icon: IconMusic },
    { label: "Poll", value: "poll", icon: IconActivity },
    { label: "Shared", value: "shared", icon: IconShare },
  ],

  visibility: [
    { label: "All Visibility", value: "all", icon: IconEye },
    { label: "Public", value: "public", icon: IconGlobe },
    { label: "Friends", value: "friends", icon: IconUsers },
    { label: "Private", value: "private", icon: IconLock },
  ],

  status: [
    { label: "All Posts", value: "all", icon: IconEye },
    { label: "Active", value: "active", icon: IconEye },
    { label: "Hidden", value: "hidden", icon: IconEyeOff },
    { label: "Reported", value: "reported", icon: IconFlag },
    { label: "Pinned", value: "pinned", icon: IconTrendingUp },
  ],
};

export default PostTableFormatters;
