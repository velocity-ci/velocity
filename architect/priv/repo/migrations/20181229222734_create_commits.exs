defmodule Architect.Repo.Migrations.CreateCommits do
  use Ecto.Migration

  def change do
    create table(:commits, primary_key: false) do
      add(:id, :uuid, primary_key: true)

      add(
        :project_id,
        references(:projects, on_delete: :delete_all, type: :uuid)
      )

      add(:hash, :string)
      add(:message, :text)
      add(:author_email, :string)

      timestamps()
    end
  end
end
