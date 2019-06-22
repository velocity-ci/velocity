defmodule Architect.Repo.Migrations.AddTasks do
  use Ecto.Migration

  def change do
    create table(:tasks, primary_key: false) do
      add(:id, :uuid, primary_key: true)
      add(:build_id, references(:builds, type: :uuid))

      add(:plan, :map)

      add(:status, :string)
      add(:created_at, :utc_datetime)
      add(:updated_at, :utc_datetime)
      add(:started_at, :utc_datetime)
      add(:completed_at, :utc_datetime)
    end
  end
end
