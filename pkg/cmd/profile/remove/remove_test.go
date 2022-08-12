package remove

import (
	"testing"

	"github.com/algolia/cli/pkg/cmdutil"
	"github.com/algolia/cli/pkg/config"
	"github.com/algolia/cli/pkg/iostreams"
	"github.com/algolia/cli/test"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRemoveCmd(t *testing.T) {
	tests := []struct {
		name      string
		tty       bool
		cli       string
		wantsErr  bool
		wantsOpts RemoveOptions
	}{
		{
			name:     "no --confirm without tty",
			cli:      "default",
			tty:      false,
			wantsErr: true,
			wantsOpts: RemoveOptions{
				DoConfirm: true,
				Profile:   "default",
			},
		},
		{
			name:     "--confirm without tty",
			cli:      "default --confirm",
			tty:      false,
			wantsErr: false,
			wantsOpts: RemoveOptions{
				DoConfirm: false,
				Profile:   "default",
			},
		},
		{
			name:     "non-existant profile",
			cli:      "foo",
			tty:      true,
			wantsErr: true,
			wantsOpts: RemoveOptions{
				DoConfirm: true,
				Profile:   "foo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, _, _ := iostreams.Test()
			if tt.tty {
				io.SetStdinTTY(tt.tty)
				io.SetStdoutTTY(tt.tty)
			}

			f := &cmdutil.Factory{
				IOStreams: io,
				Config:    test.NewDefaultConfigStub(),
			}

			var opts *RemoveOptions
			cmd := NewRemoveCmd(f, func(o *RemoveOptions) error {
				opts = o
				return nil
			})

			args, err := shlex.Split(tt.cli)
			require.NoError(t, err)
			cmd.SetArgs(args)
			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				assert.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantsOpts.Profile, opts.Profile)
			assert.Equal(t, tt.wantsOpts.DoConfirm, opts.DoConfirm)
		})
	}
}

func Test_runRemoveCmd(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		profiles map[string]bool
		wantsErr string
		wantOut  string
	}{
		{
			name:     "existing profile (default)",
			cli:      "default --confirm",
			profiles: map[string]bool{"default": true, "foo": false},
			wantOut:  "✓ 'default' removed successfully. Set a new default profile with 'algolia profile setdefault'.\n",
		},
		{
			name:     "existing profile (non-default)",
			cli:      "foo --confirm",
			profiles: map[string]bool{"default": true, "foo": false},
			wantOut:  "✓ 'foo' removed successfully.\n",
		},
		{
			name:     "non-existant profile",
			cli:      "bar --confirm",
			profiles: map[string]bool{"default": true, "foo": false},
			wantsErr: "the specified profile does not exist: 'bar'",
		},
		{
			name:     "only one profile",
			cli:      "default --confirm",
			profiles: map[string]bool{"default": true},
			wantOut:  "✓ 'default' removed successfully. Add a profile with 'algolia profile add'.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p []*config.Profile
			for k, v := range tt.profiles {
				p = append(p, &config.Profile{
					Name:    k,
					Default: v,
				})
			}
			cfg := test.NewConfigStubWithProfiles(p)
			f, out := test.NewFactory(true, nil, cfg, "")
			cmd := NewRemoveCmd(f, nil)
			out, err := test.Execute(cmd, tt.cli, out)
			if err != nil {
				assert.Equal(t, tt.wantsErr, err.Error())
				return
			}

			assert.Equal(t, tt.wantOut, out.String())
		})
	}
}