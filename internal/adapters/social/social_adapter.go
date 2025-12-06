package social

import (
	"errors"
	"regexp"

	"github.com/elchemista/driplnk/internal/config"
	"github.com/elchemista/driplnk/internal/domain"
)

type SocialAdapter struct {
	platforms []platformRule
}

type platformRule struct {
	regex    *regexp.Regexp
	platform domain.SocialPlatform
}

func NewSocialAdapter(configs []config.SocialPlatformConfig) *SocialAdapter {
	var rules []platformRule
	for _, cfg := range configs {
		regex, err := regexp.Compile(cfg.RegexPattern)
		if err != nil {
			// Skip invalid regex, maybe log it
			continue
		}
		rules = append(rules, platformRule{
			regex: regex,
			platform: domain.SocialPlatform{
				Name:    cfg.Name,
				Domain:  cfg.Domain,
				IconSVG: cfg.IconSVG,
				Color:   cfg.Color,
			},
		})
	}
	return &SocialAdapter{
		platforms: rules,
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
