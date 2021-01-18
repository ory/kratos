package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/stringslice"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"sort"
	"strings"
)

type (
	Type  string
	Group string
)

const (
	DefaultGroup          Group = "default"
	PasswordGroup         Group = "authenticator_password"
	OpenIDConnectGroup    Group = "authenticator_oidc"
	RecoveryLinkGroup     Group = "recovery_link"
	VerificationLinkGroup Group = "verification_link"

	Text   Type = "text"
	Input  Type = "input"
	Image  Type = "img"
	Anchor Type = "a"
)

type (
	Nodes []*Node

	// Node represents a flow's nodes
	//
	// Nodes are represented as HTML elements or their native UI equivalents. For example,
	// a node can be an `<img>` tag, or an `<input element>` but also `some plain text`.
	Node struct {
		// The node's type
		//
		// Can be one of: text, input, img, a
		//
		// required: true
		Type Type `json:"type" faker:"-"`

		// Group specifies which group (e.g. password authenticator) this node belongs to.
		Group Group `json:"group"`

		// The node's attributes.
		//
		// required: true
		Attributes Attributes `json:"attributes" faker:"ui_node_attributes"`

		// The node's messages
		//
		// Contains error, validation, or other messages relevant to this node.
		Messages text.Messages `json:"messages"`
	}

	// Used for en/decoding the Attributes field.
	jsonRawNode struct {
		Type       Type          `json:"type"`
		Group      Group         `json:"group"`
		Attributes Attributes    `json:"attributes"`
		Messages   text.Messages `json:"messages"`
	}
)

func (n *Node) ID() string {
	return string(n.Group) + "/" + n.Attributes.ID()
}

func (n *Node) Reset() {
	n.Messages = nil
	n.Attributes.Reset()
}

func (n *Node) GetValue() interface{} {
	return n.Attributes.GetValue()
}

func (n Nodes) Find(group Group, id string) *Node {
	for _, nn := range n {
		if nn.ID() == string(group)+"/"+id {
			return nn
		}
	}

	return nil
}

func (n Nodes) Reset(exclude ...string) {
	for k, nn := range n {
		nn.Messages = nil

		if !stringslice.Has(exclude, nn.ID()) {
			nn.Reset()
		}
		n[k] = nn
	}
}

func (n Nodes) SortBySchema(schemaRef, prefix string) error {
	schemaKeys, err := schema.GetKeysInOrder(schemaRef)
	if err != nil {
		return err
	}

	keysInOrder := []string{
		x.CSRFTokenName,
		"identifier",
		"password",
	}

	for _, k := range schemaKeys {
		if prefix != "" {
			k = fmt.Sprintf("%s.%s", prefix, k)
		}
		keysInOrder = append(keysInOrder, k)
	}

	getKeyPosition := func(name string) int {
		lastPrefix := len(keysInOrder)
		for i, n := range keysInOrder {
			if strings.HasPrefix(name, n) {
				lastPrefix = i
			}
		}
		return lastPrefix
	}

	sort.SliceStable(n, func(i, j int) bool {
		a := strings.Split(n[i].ID(), "/")
		b := strings.Split(n[j].ID(), "/")
		return getKeyPosition(a[len(a)-1]) < getKeyPosition(b[len(b)-1])
	})

	return nil
}

// Remove removes one or more nodes by their IDs.
func (n *Nodes) Remove(ids ...string) {
	if n == nil {
		return
	}

	for _, needle := range ids {
		for i := range *n {
			if (*n)[i].ID() == needle {
				*n = append((*n)[:i], (*n)[i+1:]...)
			}
		}
	}
}

// Upsert updates or appends a node.
func (n *Nodes) Upsert(node *Node) {
	if n == nil {
		*n = append(*n, node)
		return
	}

	for i := range *n {
		if (*n)[i].ID() == node.ID() {
			(*n)[i] = node
			return
		}
	}

	*n = append(*n, node)
}

// Append appends a node.
func (n *Nodes) Append(node *Node) {
	*n = append(*n, node)
}

func (n *Node) UnmarshalJSON(data []byte) error {
	var attr Attributes
	switch t := gjson.GetBytes(data, "type").String(); Type(t) {
	case Text:
		attr = new(TextAttributes)
	case Input:
		attr = new(InputAttributes)
	case Anchor:
		attr = new(AnchorAttributes)
	case Image:
		attr = new(ImageAttributes)
	default:
		return fmt.Errorf("unexpected node type: %s", t)
	}

	var d jsonRawNode
	d.Attributes = attr
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&d); err != nil {
		return err
	}

	*n = Node(d)
	return nil
}

func (n *Node) MarshalJSON() ([]byte, error) {
	var t Type
	if n.Attributes != nil {
		switch n.Attributes.(type) {
		case *TextAttributes:
			t = Text
		case *InputAttributes:
			t = Input
		case *AnchorAttributes:
			t = Anchor
		case *ImageAttributes:
			t = Image
		default:
			return nil, errors.WithStack(fmt.Errorf("unknown node type: %T", n.Attributes))
		}
	}

	if n.Type == "" {
		n.Type = t
	} else if n.Type != t {
		return nil, errors.WithStack(fmt.Errorf("node type and node attributes mismatch: %T != %s", n.Attributes, n.Type))
	}

	return json.Marshal((*jsonRawNode)(n))
}
