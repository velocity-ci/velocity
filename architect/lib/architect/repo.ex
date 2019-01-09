defmodule Architect.Repo do
  use Ecto.Repo,
    otp_app: :architect,
    adapter: Ecto.Adapters.Postgres
end
