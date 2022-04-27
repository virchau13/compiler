package transform

import (
	astro "github.com/withastro/compiler/internal"
	"golang.org/x/net/html/atom"
)

func ScopeElement(n *astro.Node, opts TransformOptions) {
	if n.Type == astro.ElementNode {
		if _, noScope := NeverScopedElements[n.Data]; !noScope {
			injectScopedClass(n, opts)
		}
	}
}

func AnnotateElement(n *astro.Node, opts TransformOptions) {
	if n.Type == astro.ElementNode && !n.Component && !n.Fragment {
		if _, noScope := NeverScopedElements[n.Data]; !noScope {
			annotateElement(n, opts)
		}
	}
}

var NeverScopedElements map[string]bool = map[string]bool{
	// "html" is a notable omission, see `NeverScopedSelectors`
	"Fragment": true,
	"base":     true,
	"body":     true,
	"font":     true,
	"frame":    true,
	"frameset": true,
	"head":     true,
	"link":     true,
	"meta":     true,
	"noframes": true,
	"noscript": true,
	"script":   true,
	"style":    true,
	"title":    true,
}

var NeverScopedSelectors map[string]bool = map[string]bool{
	// html is never scoped as a selector (from CSS) but is scoped as an element (from HTML)
	"html":  true,
	":root": true,
}

func annotateElement(n *astro.Node, opts TransformOptions) {
	if n.DataAtom == atom.Html {
		n.Attr = append(n.Attr, astro.Attribute{
			Key:  "data-astro-source-root",
			Type: astro.QuotedAttribute,
			Val:  opts.ProjectRoot,
		})
	} else {
		n.Attr = append(n.Attr, astro.Attribute{
			Key:  "data-astro-source-file",
			Type: astro.QuotedAttribute,
			Val:  opts.Pathname,
		})
	}
}

func injectScopedClass(n *astro.Node, opts TransformOptions) {
	for i, attr := range n.Attr {
		// If we find an existing class attribute, append the scoped class
		if attr.Key == "class" || (n.Component && attr.Key == "className") {
			switch attr.Type {
			case astro.ShorthandAttribute:
				if n.Component {
					attr.Val = attr.Key + ` + " astro-` + opts.Scope + `"`
					attr.Type = astro.ExpressionAttribute
					n.Attr[i] = attr
					return
				}
			case astro.EmptyAttribute:
				// instead of an empty string
				attr.Type = astro.QuotedAttribute
				attr.Val = "astro-" + opts.Scope
				n.Attr[i] = attr
				return
			case astro.QuotedAttribute, astro.TemplateLiteralAttribute:
				// as a plain string
				attr.Val = attr.Val + " astro-" + opts.Scope
				n.Attr[i] = attr
				return
			case astro.ExpressionAttribute:
				// as an expression
				attr.Val = "(" + attr.Val + `) + " astro-` + opts.Scope + `"`
				n.Attr[i] = attr
				return
			}
		}
	}
	// If we didn't find an existing class attribute, let's add one
	n.Attr = append(n.Attr, astro.Attribute{
		Key: "class",
		Val: "astro-" + opts.Scope,
	})
}
