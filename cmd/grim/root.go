package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"grimmoir/internal/clipboard"
	"grimmoir/internal/config"
	"grimmoir/internal/git"
	"grimmoir/internal/model"
	"grimmoir/internal/service"
	"grimmoir/internal/store"
	"grimmoir/internal/tui"
)

func NewRootCmd() *cobra.Command {
	var rootPath string
	var rootCmd = &cobra.Command{
		Use:   "grim",
		Short: "Manage Grimmoir prompts and skills",
	}

	rootCmd.PersistentFlags().StringVar(&rootPath, "path", config.StorePath(), "Skill store path")

	newSvc := func() (*service.Service, error) {
		repo, err := store.New(rootPath)
		if err != nil {
			return nil, err
		}
		return service.New(repo, clipboard.New()), nil
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newSvc()
			if err != nil {
				return err
			}
			skills, err := svc.List()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tVERSION\tTAGS\tDESCRIPTION")
			for _, s := range skills {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, s.Version, strings.Join(s.Tags, ","), s.Description)
			}
			return w.Flush()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "get <name>",
		Short: "Output raw prompt text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newSvc()
			if err != nil {
				return err
			}
			skill, err := svc.GetByName(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), skill.Body)
			return err
		},
	})

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Create a new skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newSvc()
			if err != nil {
				return err
			}
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("description")
			tagsRaw, _ := cmd.Flags().GetString("tags")
			version, _ := cmd.Flags().GetString("version")
			fromClip, _ := cmd.Flags().GetBool("clip")
			body, _ := cmd.Flags().GetString("body")

			skill := model.Skill{
				Name:        name,
				Description: desc,
				Tags:        parseTags(tagsRaw),
				Version:     version,
				Body:        body,
			}

			var path string
			if fromClip {
				path, err = svc.AddFromClipboard(skill)
			} else {
				path, err = svc.AddFromText(skill)
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Saved %s\n", path)
			return err
		},
	}
	addCmd.Flags().String("name", "", "Skill name")
	addCmd.Flags().String("description", "", "Skill description")
	addCmd.Flags().String("tags", "", "Comma-separated tags")
	addCmd.Flags().String("version", "0.1.0", "Skill version")
	addCmd.Flags().Bool("clip", false, "Use clipboard as body")
	addCmd.Flags().String("body", "", "Skill body text")
	addCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(addCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a skill by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newSvc()
			if err != nil {
				return err
			}
			path, err := svc.DeleteByName(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s\n", path)
			return err
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Sync prompt markdown files in store repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := git.NewSyncer(git.NewExecRunner())
			if err := s.Sync(rootPath); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Sync complete")
			return err
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Launch interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newSvc()
			if err != nil {
				return err
			}
			return tui.New(svc).Run()
		},
	})

	return rootCmd
}

func parseTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
