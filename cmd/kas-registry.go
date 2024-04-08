package cmd

import (
	"fmt"

	"github.com/opentdf/platform/protocol/go/kasregistry"
	"github.com/opentdf/tructl/pkg/cli"
	"github.com/opentdf/tructl/pkg/man"
	"github.com/spf13/cobra"
)

var (
	kasRegistry_crudCommands = []string{
		kasRegistrysCreateCmd.Use,
		kasRegistryGetCmd.Use,
		kasRegistrysListCmd.Use,
		kasRegistryUpdateCmd.Use,
		kasRegistryDeleteCmd.Use,
	}

	// KasRegistryCmd is the command for managing KAS registrations
	kasRegistryCmd = &cobra.Command{
		Use:   man.Docs.GetDoc("policy/kas-registry").Use,
		Short: man.Docs.GetDoc("policy/kas-registry").GetShort(kasRegistry_crudCommands),
		Long:  man.Docs.GetDoc("policy/kas-registry").Long,
	}

	kasRegistryGetCmd = &cobra.Command{
		Use:   man.Docs.GetDoc("policy/kas-registry/get").Use,
		Short: man.Docs.GetDoc("policy/kas-registry/get").Short,
		Run: func(cmd *cobra.Command, args []string) {
			h := cli.NewHandler(cmd)
			defer h.Close()

			flagHelper := cli.NewFlagHelper(cmd)
			id := flagHelper.GetRequiredString("id")

			kas, err := h.GetKasRegistryEntry(id)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to get KAS registry entry (%s)", id)
				cli.ExitWithError(errMsg, err)
			}

			keyType := "Local"
			key := kas.PublicKey.GetLocal()
			if kas.PublicKey.GetRemote() != "" {
				keyType = "Remote"
				key = kas.PublicKey.GetRemote()
			}

			t := cli.NewTabular().
				Rows([][]string{
					{"Id", kas.Id},
					// TODO: render labels [https://github.com/opentdf/tructl/issues/73]
					{"URI", kas.Uri},
					{"PublicKey Type", keyType},
					{"PublicKey", key},
				}...)
			HandleSuccess(cmd, kas.Id, t, kas)
		},
	}

	kasRegistrysListCmd = &cobra.Command{
		Use:   man.Docs.GetDoc("policy/kas-registry/list").Use,
		Short: man.Docs.GetDoc("policy/kas-registry/list").Short,
		Run: func(cmd *cobra.Command, args []string) {
			h := cli.NewHandler(cmd)
			defer h.Close()

			list, err := h.ListKasRegistryEntries()
			if err != nil {
				cli.ExitWithError("Failed to list KAS registry entries", err)
			}

			t := cli.NewTable()
			t.Headers("Id", "URI", "PublicKey Location", "PublicKey")
			for _, kas := range list {
				keyType := "Local"
				key := kas.PublicKey.GetLocal()
				if kas.PublicKey.GetRemote() != "" {
					keyType = "Remote"
					key = kas.PublicKey.GetRemote()
				}

				t.Row(
					kas.Id,
					kas.Uri,
					keyType,
					key,
					// TODO: render labels [https://github.com/opentdf/tructl/issues/73]
				)
			}
			HandleSuccess(cmd, "", t, list)
		},
	}

	kasRegistrysCreateCmd = &cobra.Command{
		Use:   man.Docs.GetDoc("policy/kas-registry/create").Use,
		Short: man.Docs.GetDoc("policy/kas-registry/create").Short,
		Run: func(cmd *cobra.Command, args []string) {
			h := cli.NewHandler(cmd)
			defer h.Close()

			flagHelper := cli.NewFlagHelper(cmd)
			uri := flagHelper.GetRequiredString("uri")
			local := flagHelper.GetOptionalString("public-key-local")
			remote := flagHelper.GetOptionalString("public-key-remote")
			metadataLabels := flagHelper.GetStringSlice("label", metadataLabels, cli.FlagHelperStringSliceOptions{Min: 0})

			if local == "" && remote == "" {
				e := fmt.Errorf("A public key is required. Please pass either a local or remote public key")
				cli.ExitWithError("Issue with create flags 'public-key-local' and 'public-key-remote': ", e)
			}

			key := &kasregistry.PublicKey{}
			keyType := "Local"
			if local != "" {
				if remote != "" {
					e := fmt.Errorf("Only one public key is allowed. Please pass either a local or remote public key but not both")
					cli.ExitWithError("Issue with create flags 'public-key-local' and 'public-key-remote': ", e)
				}
				key.PublicKey = &kasregistry.PublicKey_Local{Local: local}
			} else {
				keyType = "Remote"
				key.PublicKey = &kasregistry.PublicKey_Remote{Remote: remote}
			}

			created, err := h.CreateKasRegistryEntry(
				uri,
				key,
				getMetadataMutable(metadataLabels),
			)
			if err != nil {
				cli.ExitWithError("Failed to create KAS registry entry", err)
			}

			t := cli.NewTabular().
				Rows([][]string{
					{"Id", created.Id},
					{"URI", created.Uri},
					{"PublicKey Type", keyType},
					{"PublicKey", local},
					// TODO: render labels [https://github.com/opentdf/tructl/issues/73]
				}...)

			HandleSuccess(cmd, created.Id, t, created)
		},
	}

	// Update one KAS registry entry
	kasRegistryUpdateCmd = &cobra.Command{
		Use:   man.Docs.GetDoc("policy/kas-registry/update").Use,
		Short: man.Docs.GetDoc("policy/kas-registry/update").Short,
		Run: func(cmd *cobra.Command, args []string) {
			h := cli.NewHandler(cmd)
			defer h.Close()

			flagHelper := cli.NewFlagHelper(cmd)

			id := flagHelper.GetRequiredString("id")
			uri := flagHelper.GetOptionalString("uri")
			local := flagHelper.GetOptionalString("public-key-local")
			remote := flagHelper.GetOptionalString("public-key-remote")
			labels := flagHelper.GetStringSlice("label", metadataLabels, cli.FlagHelperStringSliceOptions{Min: 0})

			if local == "" && remote == "" && len(labels) == 0 && uri == "" {
				cli.ExitWithError("No values were passed to update. Please pass at least one value to update (E.G. 'uri', 'public-key-local', 'public-key-remote', 'label')", nil)
			}

			// TODO: should update of a type of key be a dangerous mutation or cause a need for confirmation in the CLI?
			var pubKey *kasregistry.PublicKey
			if local != "" && remote != "" {
				e := fmt.Errorf("Only one public key is allowed. Please pass either a local or remote public key but not both")
				cli.ExitWithError("Issue with update flags 'public-key-local' and 'public-key-remote': ", e)
			} else if local != "" {
				pubKey = &kasregistry.PublicKey{PublicKey: &kasregistry.PublicKey_Local{Local: local}}
			} else if remote != "" {
				pubKey = &kasregistry.PublicKey{PublicKey: &kasregistry.PublicKey_Remote{Remote: remote}}
			}

			updated, err := h.UpdateKasRegistryEntry(
				id,
				uri,
				pubKey,
				getMetadataMutable(labels),
				getMetadataUpdateBehavior(),
			)
			if err != nil {
				cli.ExitWithError(fmt.Sprintf("Failed to update KAS registry entry (%s)", id), err)
			}
			t := cli.NewTabular().
				Rows([][]string{
					{"Id", id},
					{"URI", uri},
					// TODO: render labels [https://github.com/opentdf/tructl/issues/73]
				}...)
			HandleSuccess(cmd, id, t, updated)
		},
	}

	kasRegistryDeleteCmd = &cobra.Command{
		Use:   man.Docs.GetDoc("policy/kas-registry/delete").Use,
		Short: man.Docs.GetDoc("policy/kas-registry/delete").Short,
		Run: func(cmd *cobra.Command, args []string) {
			h := cli.NewHandler(cmd)
			defer h.Close()

			flagHelper := cli.NewFlagHelper(cmd)
			id := flagHelper.GetRequiredString("id")

			kas, err := h.GetKasRegistryEntry(id)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to get KAS registry entry (%s)", id)
				cli.ExitWithError(errMsg, err)
			}

			cli.ConfirmAction(cli.ActionDelete, "KAS Registry Entry: ", id)

			if _, err := h.DeleteKasRegistryEntry(id); err != nil {
				errMsg := fmt.Sprintf("Failed to delete KAS registry entry (%s)", id)
				cli.ExitWithError(errMsg, err)
			}

			t := cli.NewTabular().
				Rows([][]string{
					{"Id", kas.Id},
					{"URI", kas.Uri},
				}...)

			HandleSuccess(cmd, kas.Id, t, kas)
		},
	}
)

func init() {
	policyCmd.AddCommand(kasRegistryCmd)

	getDoc := man.Docs.GetDoc("policy/kas-registry/get")
	kasRegistryCmd.AddCommand(kasRegistryGetCmd)
	kasRegistryGetCmd.Flags().StringP(
		getDoc.GetDocFlag("id").Name,
		getDoc.GetDocFlag("id").Shorthand,
		getDoc.GetDocFlag("id").Default,
		getDoc.GetDocFlag("id").Description,
	)

	kasRegistryCmd.AddCommand(kasRegistrysListCmd)
	// TODO: active, inactive, any state querying [https://github.com/opentdf/tructl/issues/68]

	createDoc := man.Docs.GetDoc("policy/kas-registry/create")
	kasRegistryCmd.AddCommand(kasRegistrysCreateCmd)
	kasRegistrysCreateCmd.Flags().StringP(
		createDoc.GetDocFlag("uri").Name,
		createDoc.GetDocFlag("uri").Shorthand,
		createDoc.GetDocFlag("uri").Default,
		createDoc.GetDocFlag("uri").Description,
	)
	kasRegistrysCreateCmd.Flags().StringP(
		createDoc.GetDocFlag("public-key-local").Name,
		createDoc.GetDocFlag("public-key-local").Shorthand,
		createDoc.GetDocFlag("public-key-local").Default,
		createDoc.GetDocFlag("public-key-local").Description,
	)
	kasRegistrysCreateCmd.Flags().StringP(
		createDoc.GetDocFlag("public-key-remote").Name,
		createDoc.GetDocFlag("public-key-remote").Shorthand,
		createDoc.GetDocFlag("public-key-remote").Default,
		createDoc.GetDocFlag("public-key-remote").Description,
	)
	injectLabelFlags(kasRegistrysCreateCmd, false)

	updateDoc := man.Docs.GetDoc("policy/kas-registry/update")
	kasRegistryCmd.AddCommand(kasRegistryUpdateCmd)
	kasRegistryUpdateCmd.Flags().StringP(
		updateDoc.GetDocFlag("id").Name,
		updateDoc.GetDocFlag("id").Shorthand,
		updateDoc.GetDocFlag("id").Default,
		updateDoc.GetDocFlag("id").Description,
	)
	kasRegistryUpdateCmd.Flags().StringP(
		updateDoc.GetDocFlag("uri").Name,
		updateDoc.GetDocFlag("uri").Shorthand,
		updateDoc.GetDocFlag("uri").Default,
		updateDoc.GetDocFlag("uri").Description,
	)
	kasRegistryUpdateCmd.Flags().StringP(
		updateDoc.GetDocFlag("public-key-local").Name,
		updateDoc.GetDocFlag("public-key-local").Shorthand,
		updateDoc.GetDocFlag("public-key-local").Default,
		updateDoc.GetDocFlag("public-key-local").Description,
	)
	kasRegistryUpdateCmd.Flags().StringP(
		updateDoc.GetDocFlag("public-key-remote").Name,
		updateDoc.GetDocFlag("public-key-remote").Shorthand,
		updateDoc.GetDocFlag("public-key-remote").Default,
		updateDoc.GetDocFlag("public-key-remote").Description,
	)
	injectLabelFlags(kasRegistryUpdateCmd, true)

	deleteDoc := man.Docs.GetDoc("policy/kas-registry/delete")
	kasRegistryCmd.AddCommand(kasRegistryDeleteCmd)
	kasRegistryDeleteCmd.Flags().StringP(
		deleteDoc.GetDocFlag("id").Name,
		deleteDoc.GetDocFlag("id").Shorthand,
		deleteDoc.GetDocFlag("id").Default,
		deleteDoc.GetDocFlag("id").Description,
	)
}

func init() {
	rootCmd.AddCommand(kasRegistryCmd)
}
