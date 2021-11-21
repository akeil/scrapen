package content

import (
	"bytes"
	_ "embed"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
)

type ruleset int

const (
	rulesPrep ruleset = iota
	rulesClean
)

const (
	actDrop   = "drop"
	actUnwrap = "unwrap"
)

//go:embed rules-prepare.yaml
var prepRules []byte

func applyRules(rs ruleset, doc *goquery.Document) {
	var data []byte
	switch rs {
	case rulesPrep:
		data = prepRules
	default:
		log.WithFields(log.Fields{
			"ruleset": rs,
			"module":  "rules",
		}).Warning("Rules will not be applied, unknown ruleset")
		return
	}

	rules, err := loadRules(data)
	if err != nil {
		log.WithFields(log.Fields{
			"module":  "rules",
			"ruleset": rs,
			"err":     err,
		}).Warning("Failed to load ruleset")
		return
	}

	log.WithFields(log.Fields{
		"ruleset": rs,
		"module":  "rules",
		"count":   len(rules),
	}).Info("Apply rules")

	for _, rule := range rules {
		rule.Apply(doc)
	}
}

// loadRules parses rules from the given data in YAML format
func loadRules(data []byte) ([]rule, error) {
	r := bytes.NewReader(data)
	dec := yaml.NewDecoder(r)
	rules := make([]*configRule, 0)
	err := dec.Decode(&rules)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		rule.compile()
	}

	// not sure if this is the best way to convert to interface
	x := make([]rule, len(rules))
	for i, y := range rules {
		x[i] = y
	}
	return x, nil
}

type rule interface {
	Apply(*goquery.Document)
}

type configRule struct {
	Action     string
	Elements   []string
	Attr       string
	Values     []string
	valRegex   []*regexp.Regexp
	Classes    []string
	classRegex []*regexp.Regexp
}

func (c *configRule) compile() {
	c.classRegex = make([]*regexp.Regexp, 0)
	for _, s := range c.Classes {
		re, err := regexp.Compile(s)
		if err != nil {
			//TODO: log error
		} else {
			c.classRegex = append(c.classRegex, re)
		}
	}

	c.valRegex = make([]*regexp.Regexp, 0)
	for _, s := range c.Values {
		re, err := regexp.Compile(s)
		if err != nil {
			//TODO: log error
		} else {
			c.valRegex = append(c.valRegex, re)
		}
	}
}

func (c *configRule) Apply(doc *goquery.Document) {
	// Select affected elements
	var tags string
	if len(c.Elements) != 0 {
		tags = strings.Join(c.Elements, ",")
	} else {
		tags = "*"
	}
	s := doc.Find(tags)

	if c.Attr != "" {
		c.applyForAttr(s)
	}

	if len(c.Classes) != 0 {
		c.applyForClasses(s)
	}

	if c.Attr == "" && len(c.Classes) == 0 {
		// if we have no further restrictions, apply Action on the selected elements
		log.WithFields(log.Fields{
			"action":   c.Action,
			"elements": tags,
			"affected": s.Size(),
		}).Info("Apply for elements")

		c.doApply(s)
	}
}

func (c configRule) applyForAttr(s *goquery.Selection) {
	s.Each(func(i int, e *goquery.Selection) {
		v, exists := e.Attr(c.Attr)
		if !exists {
			return
		}

		for _, re := range c.valRegex {
			if re.MatchString(v) {
				log.WithFields(log.Fields{
					"action":    c.Action,
					"attribute": c.Attr,
					"value":     v,
					"matches":   re.String(),
					"affected":  s.Size(),
				}).Info("Apply for attribute")

				c.doApply(s)
				return
			}
		}
	})
}

func (c configRule) applyForClasses(s *goquery.Selection) {
	s.Each(func(i int, e *goquery.Selection) {
		v, exists := e.Attr("class")
		if !exists {
			return
		}

		tag := goquery.NodeName(e)
		if tag == "html" || tag == "body" || tag == "article" {
			return
		}

		elemClasses := strings.Split(v, " ")
		for _, cls := range elemClasses {
			for _, re := range c.classRegex {
				if re.MatchString(cls) {
					log.WithFields(log.Fields{
						"action":  c.Action,
						"tag":     tag,
						"class":   cls,
						"matches": re.String(),
					}).Info("Apply for class")
					c.doApply(e)
					return
				}
			}
		}
	})
}

func (c configRule) doApply(s *goquery.Selection) {
	switch c.Action {
	case actDrop:
		s.Remove()
		break
	case actUnwrap:
		if s.Contents().Length() == 0 {
			// cannot unwrap if empty
			s.Remove()
		} else {
			s.Contents().Unwrap()
		}
		break
	}
}
