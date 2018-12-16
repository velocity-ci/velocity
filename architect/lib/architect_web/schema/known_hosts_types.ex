defmodule ArchitectWeb.Schema.KnownHostsTypes do
  use Absinthe.Schema.Notation

  object :known_host do
    field(:id, :id)
    field(:comment, :string)
    field(:entry, :string)
    field(:fingerprint_md5, :string)
    field(:fingerprint_sha256, :string)
    field(:hosts, :string)
  end
end
