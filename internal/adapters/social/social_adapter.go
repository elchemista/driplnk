package social

import (
	"errors"
	"regexp"

	"github.com/elchemista/driplnk/internal/domain"
)

type SocialAdapter struct {
	platforms []platformRule
}

type platformRule struct {
	regex    *regexp.Regexp
	platform domain.SocialPlatform
}

func NewSocialAdapter() *SocialAdapter {
	return &SocialAdapter{
		platforms: []platformRule{
			{
				regex: regexp.MustCompile(`(?i)(facebook\.com|fb\.com)`),
				platform: domain.SocialPlatform{
					Name:    "Facebook",
					Domain:  "facebook.com",
					IconSVG: "facebook_icon_svg_here", // Placeholder
					Color:   "#1877F2",
				},
			},
			{
				regex: regexp.MustCompile(`(?i)(twitter\.com|x\.com)`),
				platform: domain.SocialPlatform{
					Name:    "Twitter",
					Domain:  "twitter.com",
					IconSVG: "twitter_icon_svg_here", // Placeholder
					Color:   "#1DA1F2",
				},
			},
			{
				regex: regexp.MustCompile(`(?i)(instagram\.com)`),
				platform: domain.SocialPlatform{
					Name:    "Instagram",
					Domain:  "instagram.com",
					IconSVG: "instagram_icon_svg_here", // Placeholder
					Color:   "#E1306C",
				},
			},
			{
				regex: regexp.MustCompile(`(?i)(linkedin\.com)`),
				platform: domain.SocialPlatform{
					Name:    "LinkedIn",
					Domain:  "linkedin.com",
					IconSVG: "linkedin_icon_svg_here", // Placeholder
					Color:   "#0077B5",
				},
			},
			{
				regex: regexp.MustCompile(`(?i)(youtube\.com|youtu\.be)`),
				platform: domain.SocialPlatform{
					Name:    "YouTube",
					Domain:  "youtube.com",
					IconSVG: "youtube_icon_svg_here", // Placeholder
					Color:   "#FF0000",
				},
			},
			{
				regex: regexp.MustCompile(`(?i)(github\.com)`),
				platform: domain.SocialPlatform{
					Name:    "GitHub",
					Domain:  "github.com",
					IconSVG: "github_icon_svg_here", // Placeholder
					Color:   "#181717",
				},
			},
			{
				regex: regexp.MustCompile(`(?i)(tiktok\.com)`),
				platform: domain.SocialPlatform{
					Name:    "TikTok",
					Domain:  "tiktok.com",
					IconSVG: "tiktok_icon_svg_here", // Placeholder
					Color:   "#000000",
				},
			},
		},
	}
}

func (s *SocialAdapter) Resolve(url string) (*domain.SocialPlatform, error) {
	for _, rule := range s.platforms {
		if rule.regex.MatchString(url) {
			// Return a copy to avoid modification of local instance if that ever happens
			p := rule.platform
			return &p, nil
		}
	}
	return nil, errors.New("platform not found")
}
