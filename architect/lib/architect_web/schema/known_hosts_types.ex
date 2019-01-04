defmodule ArchitectWeb.Schema.KnownHostsTypes do
  use Absinthe.Schema.Notation

  object :known_host do
    field(:id, non_null(:id))
    field(:entry, non_null(:string))
    field(:host, non_null(:string))
    field(:fingerprint_md5, non_null(:string))
    field(:fingerprint_sha256, non_null(:string))
    field(:verified, non_null(:boolean))
  end
end
