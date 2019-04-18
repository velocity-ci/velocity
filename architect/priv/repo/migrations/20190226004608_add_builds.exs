defmodule Architect.Repo.Migrations.AddBuilds do
  use Ecto.Migration

  def change do
    create table(:builds, primary_key: false) do
      add(:id, :uuid, primary_key: true)
      add(:project_id, references(:projects, type: :uuid))
      add(:commit_sha, :string)
      add(:branch_name, :string)
      add(:task_name, :string)
      add(:plan, :map)
      add(:parameters, :map)

      add(:status, :string)
      add(:created_at, :utc_datetime)
      add(:updated_at, :utc_datetime)
      add(:started_at, :utc_datetime)
      add(:completed_at, :utc_datetime)

      add(:created_by_id, references(:users, type: :uuid))
    end
  end
end
