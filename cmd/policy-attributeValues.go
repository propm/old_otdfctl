package cmd

import (
	"fmt"

	"github.com/opentdf/otdfctl/pkg/cli"
	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/spf13/cobra"
)

// TODO: add metadata to outputs once [https://github.com/opentdf/otdfctl/issues/73] is addressed

var (
	policy_attributeValuesCmd = &cobra.Command{
		Use:   "values",
		Short: "Manage attribute values",
	}

	policy_attributeValuesCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an attribute value",
		Run: func(cmd *cobra.Command, args []string) {
			flagHelper := cli.NewFlagHelper(cmd)
			attrId := flagHelper.GetRequiredString("attribute-id")
			value := flagHelper.GetRequiredString("value")
			metadataLabels := flagHelper.GetStringSlice("label", metadataLabels, cli.FlagHelperStringSliceOptions{Min: 0})
			// TODO: support create with members when update is unblocked to remove/alter them after creation [https://github.com/opentdf/platform/issues/476]

			h := cli.NewHandler(cmd)
			defer h.Close()

			attr, err := h.GetAttribute(attrId)
			if err != nil {
				cli.ExitWithError(fmt.Sprintf("Failed to get parent attribute (%s)", attrId), err)
			}

			v, err := h.CreateAttributeValue(attr.Id, value, getMetadataMutable(metadataLabels))
			if err != nil {
				cli.ExitWithError("Failed to create attribute value", err)
			}

			handleValueSuccess(cmd, v)
		},
	}

	policy_attributeValuesGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get an attribute value",
		Run: func(cmd *cobra.Command, args []string) {
			flagHelper := cli.NewFlagHelper(cmd)
			id := flagHelper.GetRequiredString("id")

			h := cli.NewHandler(cmd)
			defer h.Close()

			v, err := h.GetAttributeValue(id)
			if err != nil {
				cli.ExitWithError("Failed to find attribute value", err)
			}

			handleValueSuccess(cmd, v)
		},
	}

	policy_attributeValuesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List attribute value",
		Run: func(cmd *cobra.Command, args []string) {
			h := cli.NewHandler(cmd)
			defer h.Close()
			flagHelper := cli.NewFlagHelper(cmd)
			attrId := flagHelper.GetRequiredString("attribute-id")
			state := cli.GetState(cmd)
			vals, err := h.ListAttributeValues(attrId, state)
			if err != nil {
				cli.ExitWithError("Failed to list attribute values", err)
			}
			t := cli.NewTable()
			t.Headers("Id", "Fqn", "Members", "Active")
			for _, val := range vals {
				v := cli.GetSimpleAttributeValue(val)
				t.Row(
					v.Id,
					v.FQN,
					cli.CommaSeparated(v.Members),
					v.Active,
				)
			}
			HandleSuccess(cmd, "", t, vals)
		},
	}

	policy_attributeValuesUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update an attribute value",
		Run: func(cmd *cobra.Command, args []string) {
			flagHelper := cli.NewFlagHelper(cmd)
			id := flagHelper.GetRequiredString("id")
			metadataLabels := flagHelper.GetStringSlice("label", metadataLabels, cli.FlagHelperStringSliceOptions{Min: 0})

			h := cli.NewHandler(cmd)
			defer h.Close()

			_, err := h.GetAttributeValue(id)
			if err != nil {
				cli.ExitWithError(fmt.Sprintf("Failed to get attribute value (%s)", id), err)
			}

			v, err := h.UpdateAttributeValue(id, nil, getMetadataMutable(metadataLabels), getMetadataUpdateBehavior())
			if err != nil {
				cli.ExitWithError("Failed to update attribute value", err)
			}

			handleValueSuccess(cmd, v)
		},
	}

	policy_attributeValuesDeactivateCmd = &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate an attribute value",
		Run: func(cmd *cobra.Command, args []string) {
			flagHelper := cli.NewFlagHelper(cmd)
			id := flagHelper.GetRequiredString("id")

			h := cli.NewHandler(cmd)
			defer h.Close()

			value, err := h.GetAttributeValue(id)
			if err != nil {
				cli.ExitWithError(fmt.Sprintf("Failed to get attribute value (%s)", id), err)
			}

			cli.ConfirmAction(cli.ActionDeactivate, "attribute value", value.Value)

			deactivated, err := h.DeactivateAttributeValue(id)
			if err != nil {
				cli.ExitWithError("Failed to deactivate attribute value", err)
			}

			handleValueSuccess(cmd, deactivated)
		},
	}

	// TODO: uncomment when update with members is enabled in the platform [https://github.com/opentdf/platform/issues/476]
	///
	/// Attribute Value Members
	///
	// attrValueMembers = []string{}

	// policy_attributeValueMembersCmd = &cobra.Command{
	// 	Use:   "members",
	// 	Short: "Manage attribute value members",
	// 	Long:  "Manage attribute value members",
	// }

	// // Add member to attribute value
	// policy_attributeValueMembersAddCmd = &cobra.Command{
	// 	Use:   "add",
	// 	Short: "Add members to an attribute value",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		flagHelper := cli.NewFlagHelper(cmd)
	// 		id := flagHelper.GetRequiredString("id")
	// 		members := flagHelper.GetStringSlice("member", attrValueMembers, cli.FlagHelperStringSliceOptions{})

	// 		h := cli.NewHandler(cmd)
	// 		defer h.Close()

	// 		prev, err := h.GetAttributeValue(id)
	// 		if err != nil {
	// 			cli.ExitWithError(fmt.Sprintf("Failed to get attribute value (%s)", id), err)
	// 		}

	// 		action := fmt.Sprintf("%s [%s] to", cli.ActionMemberAdd, strings.Join(members, ", "))
	// 		cli.ConfirmAction(action, "attribute value", id)

	// 		prevMemberIds := make([]string, len(prev.Members))
	// 		for i, m := range prev.Members {
	// 			prevMemberIds[i] = m.GetId()
	// 		}
	// 		updated := append(prevMemberIds, members...)

	// 		v, err := h.UpdateAttributeValue(id, updated, nil, common.MetadataUpdateEnum_METADATA_UPDATE_ENUM_UNSPECIFIED)
	// 		if err != nil {
	// 			cli.ExitWithError(fmt.Sprintf("Failed to %s [%s] to attribute value (%s)", cli.ActionMemberAdd, strings.Join(members, ", "), id), err)
	// 		}

	// 		handleValueSuccess(cmd, v)
	// 	},
	// }

	// // Remove member from attribute value
	// policy_attributeValueMembersRemoveCmd = &cobra.Command{
	// 	Use:   "remove",
	// 	Short: "Remove members from an attribute value",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		flagHelper := cli.NewFlagHelper(cmd)
	// 		id := flagHelper.GetRequiredString("id")
	// 		members := flagHelper.GetStringSlice("members", attrValueMembers, cli.FlagHelperStringSliceOptions{})

	// 		h := cli.NewHandler(cmd)
	// 		defer h.Close()

	// 		prev, err := h.GetAttributeValue(id)
	// 		if err != nil {
	// 			cli.ExitWithError(fmt.Sprintf("Failed to get attribute value (%s)", id), err)
	// 		}

	// 		action := fmt.Sprintf("%s [%s] from", cli.ActionMemberRemove, strings.Join(members, ", "))
	// 		cli.ConfirmAction(action, "attribute value", id)

	// 		// collect the member ids off the members, then make the removals
	// 		updatedMemberIds := make([]string, len(prev.Members))
	// 		for i, m := range prev.Members {
	// 			updatedMemberIds[i] = m.GetId()
	// 		}
	// 		for _, toBeRemoved := range members {
	// 			for i, str := range updatedMemberIds {
	// 				if toBeRemoved == str {
	// 					updatedMemberIds = append(updatedMemberIds[:i], updatedMemberIds[i+1:]...)
	// 					break
	// 				}
	// 			}
	// 		}

	// 		v, err := h.UpdateAttributeValue(id, updatedMemberIds, nil, common.MetadataUpdateEnum_METADATA_UPDATE_ENUM_UNSPECIFIED)
	// 		if err != nil {
	// 			cli.ExitWithError(fmt.Sprintf("Failed to %s [%s] from attribute value (%s)", cli.ActionMemberRemove, strings.Join(members, ", "), id), err)
	// 		}

	// 		handleValueSuccess(cmd, v)
	// 	},
	// }

	// // Replace members of attribute value
	// policy_attributeValueMembersReplaceCmd = &cobra.Command{
	// 	Use:   "replace",
	// 	Short: "Replace members from an attribute value",
	// 	Long:  "This command will replace the members of an attribute value with the provided members. ",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		flagHelper := cli.NewFlagHelper(cmd)
	// 		id := flagHelper.GetRequiredString("id")
	// 		members := flagHelper.GetStringSlice("members", attrValueMembers, cli.FlagHelperStringSliceOptions{})

	// 		h := cli.NewHandler(cmd)
	// 		defer h.Close()

	// 		prev, err := h.GetAttributeValue(id)
	// 		if err != nil {
	// 			cli.ExitWithError(fmt.Sprintf("Failed to find attribute value (%s)", id), err)
	// 		}

	// 		existingMemberIds := make([]string, len(prev.Members))
	// 		for i, m := range prev.Members {
	// 			existingMemberIds[i] = m.GetId()
	// 		}

	// 		action := fmt.Sprintf("%s [%s] with [%s] under", cli.ActionMemberReplace, strings.Join(existingMemberIds, ", "), strings.Join(members, ", "))
	// 		cli.ConfirmAction(action, "attribute value", id)

	// 		v, err := h.UpdateAttributeValue(id, members, nil, common.MetadataUpdateEnum_METADATA_UPDATE_ENUM_UNSPECIFIED)
	// 		if err != nil {
	// 			cli.ExitWithError(fmt.Sprintf("Failed to %s of attribute value (%s)", cli.ActionMemberReplace, id), err)
	// 		}

	// 		handleValueSuccess(cmd, v)
	// 	},
	// }
)

func init() {
	policy_attributesCmd.AddGroup(
		&cobra.Group{
			ID:    "subcommand",
			Title: "Subcommands",
		},
	)
	policy_attributesCmd.AddCommand(policy_attributeValuesCmd)
	policy_attributeValuesCmd.GroupID = "subcommand"
	policy_attributeValuesCmd.AddGroup(
		&cobra.Group{
			ID:    "subcommand",
			Title: "Subcommands",
		},
	)

	policy_attributeValuesCmd.AddCommand(policy_attributeValuesCreateCmd)
	policy_attributeValuesCreateCmd.Flags().StringP("attribute-id", "a", "", "Attribute id")
	policy_attributeValuesCreateCmd.Flags().StringP("value", "v", "", "Value")
	injectLabelFlags(policy_attributeValuesCreateCmd, false)

	policy_attributeValuesCmd.AddCommand(policy_attributeValuesGetCmd)
	policy_attributeValuesGetCmd.Flags().StringP("id", "i", "", "Attribute value id")

	policy_attributeValuesCmd.AddCommand(policy_attributeValuesListCmd)
	policy_attributeValuesListCmd.Flags().StringP("attribute-id", "a", "", "Attribute id")
	policy_attributeValuesListCmd.Flags().StringP("state", "s", "active", "Filter by state [active, inactive, any]")

	policy_attributeValuesCmd.AddCommand(policy_attributeValuesUpdateCmd)
	policy_attributeValuesUpdateCmd.Flags().StringP("id", "i", "", "Attribute value id")
	injectLabelFlags(policy_attributeValuesUpdateCmd, true)

	policy_attributeValuesCmd.AddCommand(policy_attributeValuesDeactivateCmd)
	policy_attributeValuesDeactivateCmd.Flags().StringP("id", "i", "", "Attribute value id")

	// Attribute value members
	// policy_attributeValuesCmd.AddCommand(policy_attributeValueMembersCmd)
	// policy_attributeValueMembersCmd.GroupID = "subcommand"

	// policy_attributeValueMembersCmd.AddCommand(policy_attributeValueMembersAddCmd)
	// policy_attributeValueMembersAddCmd.Flags().StringP("id", "i", "", "Attribute value id")
	// policy_attributeValueMembersAddCmd.Flags().StringSliceVar(&attrValueMembers, "member", []string{}, "Each member id to add")

	// policy_attributeValueMembersCmd.AddCommand(policy_attributeValueMembersRemoveCmd)
	// policy_attributeValueMembersRemoveCmd.Flags().StringP("id", "i", "", "Attribute value id")
	// policy_attributeValueMembersRemoveCmd.Flags().StringSliceVar(&attrValueMembers, "member", []string{}, "Each member id to remove")

	// policy_attributeValueMembersCmd.AddCommand(policy_attributeValueMembersReplaceCmd)
	// policy_attributeValueMembersReplaceCmd.Flags().StringP("id", "i", "", "Attribute value id")
	// policy_attributeValueMembersReplaceCmd.Flags().StringSliceVar(&attrValueMembers, "member", []string{}, "Each member id that should exist after replacement")
}

func handleValueSuccess(cmd *cobra.Command, v *policy.Value) {
	rows := [][]string{
		{"Id", v.Id},
		{"FQN", v.Fqn},
		{"Value", v.Value},
	}
	members := v.GetMembers()
	if len(members) > 0 {
		memberIds := make([]string, len(members))
		for i, m := range members {
			memberIds[i] = m.Id
		}
		rows = append(rows, []string{"Members", cli.CommaSeparated(memberIds)})
	}
	t := cli.NewTabular().Rows(rows...)
	HandleSuccess(cmd, v.Id, t, v)
}
