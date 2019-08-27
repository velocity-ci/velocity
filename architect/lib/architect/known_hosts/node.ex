defmodule Architect.KnownHosts.Node do
  require Logger

  def host_present(host, entry) do
    # ssh-keygen -F <host>
    ## if exit != 0, append entry to .ssh/known_hosts and ssh-keygen -H
    ## if matches, verify only 1 then check entry key matches
    ## if more than 1, ssh-keygen -R <host> will remove all, then append entry to .ssh/known_hosts and ssh-keygen -H
    case get_host_entries(host) do
      [] ->
        Logger.debug("#{host} not found in #{known_hosts_path()}")
        add_host_entry(entry)

      [current_entry] ->
        Logger.debug("found #{host} in #{known_hosts_path()}")
        [_hashed_host, current_method, current_sha] = String.split(current_entry, " ")
        [_host, method, sha] = String.split(entry, " ")

        if method != current_method or String.trim(sha) != current_sha do
          Logger.debug("#{host} mistmatched in #{known_hosts_path()}")
          remove_host_entries(host)
          add_host_entry(entry)
        else
          Logger.debug("#{host} valid in #{known_hosts_path()}")
        end

      _entries ->
        Logger.debug("multiple entries found for #{host} in #{known_hosts_path()}")
        remove_host_entries(host)
        add_host_entry(entry)
    end
  end

  def host_absent(host) do
    remove_host_entries(host)
  end

  defp get_host_entries(host) do
    {out, status} = System.cmd("ssh-keygen", ["-F", host, "-f", known_hosts_path()])

    case status do
      0 ->
        lines = String.split(out, "\n", trim: true)

        Enum.filter(lines, fn line ->
          not String.starts_with?(line, "#")
        end)

      1 ->
        []
    end
  end

  defp remove_host_entries(host) do
    Logger.debug("removing all entries for #{host} from #{known_hosts_path()}")

    {_out, status} = System.cmd("ssh-keygen", ["-R", host, "-f", known_hosts_path()])

    case status do
      0 -> :ok
      _ -> :error
    end
  end

  defp add_host_entry(entry) do
    Logger.debug("adding #{String.trim(entry)} to #{known_hosts_path()}")
    {:ok, file} = File.open(known_hosts_path(), [:append])
    IO.binwrite(file, entry)
    File.close(file)

    {_out, status} = System.cmd("ssh-keygen", ["-H", "-f", known_hosts_path()])

    case status do
      0 -> :ok
      _ -> :error
    end
  end

  defp known_hosts_path do
    File.mkdir_p(known_hosts_path())
    {:ok, cwd} = File.cwd()
    "#{cwd}/.architect_data/.ssh/known_hosts"
  end
end
