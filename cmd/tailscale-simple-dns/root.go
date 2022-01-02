package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/txn2/txeh"
	"tailscale.com/client/tailscale"
)

type RootCmd struct {
	lgr           *logrus.Entry
	Timeout       time.Duration
	Domains       StringSlice
	CronExpersion string
	DryRun        bool
}

func NewRootCmd(out io.Writer) (*ffcli.Command, *RootCmd) {
	log := logrus.New()
	log.SetOutput(out)
	lgr := logrus.NewEntry(log).WithField("version", version)

	cmd := &RootCmd{
		lgr: lgr,
	}

	cmdName := "tailscale-simple-dns"

	fs := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       cmdName,
		ShortUsage: cmdName + " [flags]",
		FlagSet:    fs,
		Exec:       cmd.Exec,
		Options:    []ff.Option{ff.WithEnvVarNoPrefix()},
	}, cmd
}

func (cmd *RootCmd) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&cmd.CronExpersion, "cron", "@every 1m", "controls how frequently the sync runs can be any vaild cron experssion")
	fs.Var(&cmd.Domains, "domains", "required: domains to append to the tailscale hostname")
	fs.DurationVar(&cmd.Timeout, "timeout", time.Second*10, "set a timeout for the entire operation")
	fs.BoolVar(&cmd.DryRun, "dry-run", true, "dry run will print the updated hosts file to os.Stdout rather than updating /etc/hosts")
}

func (cmd *RootCmd) String() string {
	return fmt.Sprintf("domains to append %s every %s with timeout %s\n", cmd.Domains, cmd.CronExpersion, cmd.Timeout)
}

func (cmd *RootCmd) Exec(ctx context.Context, args []string) error {
	cmd.lgr.Info(cmd.String())
	if len(cmd.Domains) == 0 {
		return fmt.Errorf("-domains MUST be specified")
	}

	fn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cmd.Timeout)
		defer cancel()

		if err := cmd.runSync(ctx); err != nil {
			cmd.lgr.WithError(err).Error("error running sync")
		}
	}
	// run the function once
	fn()

	// run the function as a cron
	c := cron.New()
	if err := c.AddFunc(cmd.CronExpersion, fn); err != nil {
		return fmt.Errorf("cron expression (%s) was invalid: %w", cmd.CronExpersion, err)
	}
	// c.Run blocks
	c.Run()
	return nil
}

func (cmd *RootCmd) runSync(ctx context.Context) error {
	hosts, err := cmd.getTailscaleHosts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tailscale hosts: %w", err)
	}
	return cmd.updateHostsFile(hosts)
}

func (cmd *RootCmd) getTailscaleHosts(ctx context.Context) ([]HostEntry, error) {
	status, err := tailscale.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run tailscale status command: %w", err)
	}

	hostEntries := []HostEntry{}

	for _, peer := range status.Peer {
		hostEntry := HostEntry{
			Host:      peer.HostName,
			Addresses: []string{},
		}

		for _, ip := range peer.TailscaleIPs {
			// only add ipv4 addresses to simplify life
			// the hostfile libary we are using doesnt support adding both ipv4 and ipv6 for the same host
			if !ip.Is4() {
				continue
			}
			hostEntry.Addresses = append(hostEntry.Addresses, ip.String())
		}

		hostEntries = append(hostEntries, hostEntry)
	}
	cmd.lgr.WithField("host_entries", hostEntries).Infof("found %d hosts from tailscale", len(hostEntries))
	return hostEntries, nil
}

type HostEntry struct {
	Host      string
	Addresses []string
}

func (cmd *RootCmd) updateHostsFile(entries []HostEntry) error {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return fmt.Errorf("failed to initilize hosts: %w", err)
	}

	for _, entry := range entries {
		for _, address := range entry.Addresses {
			for _, domain := range cmd.Domains {
				host := fmt.Sprintf("%s.%s", entry.Host, domain)
				hosts.AddHost(address, host)
			}
		}
	}

	if cmd.DryRun {
		cmd.lgr.Print(hosts.RenderHostsFile())
	} else {
		if err := hosts.Save(); err != nil {
			return fmt.Errorf("failed saving hosts file: %w", err)
		}
	}

	return nil
}

// StringSlice is a flag.Value that collects each Set string
// into a slice, allowing for repeated flags.
type StringSlice []string

// Set implements flag.Value and appends the string to the slice.
func (ss *StringSlice) Set(s string) error {
	(*ss) = append(*ss, strings.Split(s, ",")...)
	return nil
}

// String implements flag.Value and returns the list of
// strings, or "..." if no strings have been added.
func (ss *StringSlice) String() string {
	if len(*ss) == 0 {
		return "..."
	}
	return strings.Join(*ss, ", ")
}
