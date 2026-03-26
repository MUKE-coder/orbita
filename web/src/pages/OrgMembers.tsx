import { useState } from "react";
import { useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2, UserPlus, Trash2, Mail } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { orgsApi } from "@/api/orgs";

const inviteSchema = z.object({
  email: z.string().email("Invalid email"),
  role: z.enum(["admin", "developer", "viewer"]),
});

type InviteForm = z.infer<typeof inviteSchema>;

const roleBadgeColor: Record<string, string> = {
  owner: "bg-amber-500/10 text-amber-500",
  admin: "bg-blue-500/10 text-blue-500",
  developer: "bg-green-500/10 text-green-500",
  viewer: "bg-gray-500/10 text-gray-400",
};

function OrgMembers() {
  const { orgSlug } = useParams<{ orgSlug: string }>();
  const queryClient = useQueryClient();
  const [inviteOpen, setInviteOpen] = useState(false);

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ["org-members", orgSlug],
    queryFn: () => orgsApi.listMembers(orgSlug!),
  });

  const { data: invitesData } = useQuery({
    queryKey: ["org-invites", orgSlug],
    queryFn: () => orgsApi.listInvites(orgSlug!),
  });

  const removeMutation = useMutation({
    mutationFn: (userId: string) => orgsApi.removeMember(orgSlug!, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["org-members", orgSlug] });
      toast.success("Member removed");
    },
    onError: () => toast.error("Failed to remove member"),
  });

  const revokeInviteMutation = useMutation({
    mutationFn: (id: string) => orgsApi.revokeInvite(orgSlug!, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["org-invites", orgSlug] });
      toast.success("Invite revoked");
    },
    onError: () => toast.error("Failed to revoke invite"),
  });

  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors },
  } = useForm<InviteForm>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { role: "developer" },
  });

  const inviteMutation = useMutation({
    mutationFn: (data: InviteForm) => orgsApi.inviteMember(orgSlug!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["org-invites", orgSlug] });
      toast.success("Invitation sent");
      reset();
      setInviteOpen(false);
    },
    onError: (err: unknown) => {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Failed to send invite";
      toast.error(message);
    },
  });

  const members = membersData?.data?.data || [];
  const invites = invitesData?.data?.data || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-foreground">Members</h2>
        <Button onClick={() => setInviteOpen(!inviteOpen)} size="sm">
          <UserPlus className="mr-2 h-4 w-4" />
          Invite Member
        </Button>
      </div>

      {inviteOpen && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Invite a team member</CardTitle>
          </CardHeader>
          <CardContent>
            <form
              onSubmit={handleSubmit((d) => inviteMutation.mutate(d))}
              className="flex gap-3 items-end"
            >
              <div className="flex-1 space-y-1">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="teammate@example.com"
                  {...register("email")}
                />
                {errors.email && (
                  <p className="text-xs text-destructive">{errors.email.message}</p>
                )}
              </div>
              <div className="w-36 space-y-1">
                <Label>Role</Label>
                <Select
                  defaultValue="developer"
                  onValueChange={(v) =>
                    setValue("role", v as "admin" | "developer" | "viewer")
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="developer">Developer</SelectItem>
                    <SelectItem value="viewer">Viewer</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Button
                type="submit"
                disabled={inviteMutation.isPending}
                size="sm"
              >
                {inviteMutation.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                Send
              </Button>
            </form>
          </CardContent>
        </Card>
      )}

      {membersLoading ? (
        <div className="flex justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="space-y-2">
          {members.map((member) => (
            <div
              key={member.user_id}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-sm font-medium">
                  {member.user?.name?.[0]?.toUpperCase() || "?"}
                </div>
                <div>
                  <p className="text-sm font-medium">{member.user?.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {member.user?.email}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge className={roleBadgeColor[member.role] || ""} variant="secondary">
                  {member.role}
                </Badge>
                {member.role !== "owner" && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground hover:text-destructive"
                    onClick={() => removeMutation.mutate(member.user_id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {invites.length > 0 && (
        <div className="space-y-2">
          <h3 className="text-sm font-medium text-muted-foreground">
            Pending Invitations
          </h3>
          {invites.map((invite) => (
            <div
              key={invite.id}
              className="flex items-center justify-between rounded-lg border border-dashed p-3"
            >
              <div className="flex items-center gap-3">
                <Mail className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm">{invite.email}</p>
                  <p className="text-xs text-muted-foreground">
                    Invited as {invite.role}
                  </p>
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="text-muted-foreground hover:text-destructive"
                onClick={() => revokeInviteMutation.mutate(invite.id)}
              >
                Revoke
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default OrgMembers;
