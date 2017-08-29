defmodule Mix.Tasks.CreateAdmin do
    use Mix.Task
    import Mix.Ecto

    @shortdoc "Simply runs the Hello.say/0 command."    
    def run(_) do
        ensure_started(Velocity.Repo, [])

        if Velocity.UserRepository.find_by_username("admin") == {:error} do
            password = :base64.encode(:crypto.strong_rand_bytes(16))
            user = Velocity.User.changeset(%Velocity.User{}, %{username: "admin", password: password})
            Velocity.User.register(user)
            IO.puts "\n\nCreated administrator:\n\t username: admin\n\t password: #{password}\n\n"
          end
    end
end