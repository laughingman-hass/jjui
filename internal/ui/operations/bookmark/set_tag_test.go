package bookmark

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/assert"
)

func TestSetTagModel_Update(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.TagListAll())
	commandRunner.Expect(jj.TagSet("revision", "name"))
	defer commandRunner.Verify()

	op := NewSetTagOperation(test.NewTestContext(commandRunner), "revision", "")
	test.SimulateModel(op, op.Init())
	test.SimulateModel(op, test.Type("name"))
	test.SimulateModel(op, func() tea.Msg { return intents.Apply{} })
}

func TestSetTagModel_Prefill(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.TagListAll())
	commandRunner.Expect(jj.TagSet("revision", "v1.0"))
	defer commandRunner.Verify()

	op := NewSetTagOperation(test.NewTestContext(commandRunner), "revision", "v1.")
	test.SimulateModel(op, op.Init())
	if got := op.name.Value(); got != "v1." {
		t.Fatalf("expected prefilled value %q, got %q", "v1.", got)
	}
	test.SimulateModel(op, test.Type("0"))
	test.SimulateModel(op, func() tea.Msg { return intents.Apply{} })
}

func TestSetTagModel_AutocompleteCyclesTagList(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.TagListAll()).SetOutput([]byte("v1.0;.;true;false;false;false;a\nv1.0;origin;true;true;false;false;a\nv2.0;origin;true;false;false;false;b\n"))
	defer commandRunner.Verify()

	op := NewSetTagOperation(test.NewTestContext(commandRunner), "revision", "")
	test.SimulateModel(op, op.Init())
	assert.Equal(t, []string{"v1.0", "v2.0"}, op.suggestions)
	test.SimulateModel(op, func() tea.Msg { return intents.AutocompleteCycle{Reverse: true} })
	assert.Equal(t, "v2.0", op.name.Value())
}

func TestSetTagModel_CancelRunsNoCommand(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.TagListAll())
	defer commandRunner.Verify()

	op := NewSetTagOperation(test.NewTestContext(commandRunner), "revision", "")
	test.SimulateModel(op, op.Init())
	test.SimulateModel(op, test.Type("name"))
	test.SimulateModel(op, func() tea.Msg { return intents.Cancel{} })
}
