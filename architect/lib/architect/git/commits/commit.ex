defmodule Architect.Git.Commits.Commit do
  use Ecto.Schema
  import Ecto.Changeset

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "commits" do
    belongs_to(:project, Architect.Projects.Project)

    field(:hash, :string)
    field(:message, :string)

    # TODO: author into separate table with FK (author can have multiple emails)
    # Tie author to user account in Architect
    field(:author_email, :string)
    # include more signing info to compare with author and their GPG keys in User
    # field(:signed, :string)

    timestamps()
  end

  @doc false
  def changeset(commit, attrs) do
    commit
    |> cast(attrs, [:hash])
  end
end
