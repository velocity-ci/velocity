defmodule ArchitectWeb.V1.KnownHostView do
  use ArchitectWeb, :view
  alias ArchitectWeb.V1.KnownHostView

  def render("index.json", %{known_hosts: known_hosts}) do
    %{data: render_many(known_hosts, KnownHostView, "known_host.json")}
  end

  def render("show.json", %{known_host: known_host}) do
    %{data: render_one(known_host, KnownHostView, "known_host.json")}
  end

  def render("known_host.json", %{known_host: known_host}) do
    %{
      id: known_host.id,
      hosts: known_host.hosts,
      comment: known_host.comment,
      fingerprint_sha256: known_host.fingerprint_sha256,
      fingerprint_md5: known_host.fingerprint_md5
    }
  end
end
