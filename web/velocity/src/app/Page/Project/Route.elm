module Page.Project.Route exposing (Route(..), routeToPieces, route)

import UrlParser as Url exposing (parseHash, s, (</>), string, oneOf, Parser)
import Data.Commit as Commit
import Data.Task as ProjectTask
import Data.Branch as Branch
import Util exposing ((=>))


type Route
    = Overview
    | Commits Branch.Name
    | Commit Commit.Hash
    | Task Commit.Hash ProjectTask.Name
    | Settings


route : Parser (Route -> b) b
route =
    oneOf
        [ Url.map Overview (s "overview")
        , Url.map Settings (s "settings")
        , Url.map Commits (s "commits" </> Branch.nameParser)
        , Url.map Commit (s "commit" </> Commit.hashParser)
        , Url.map Task (s "commit" </> Commit.hashParser </> s "tasks" </> ProjectTask.nameParser)
        ]



-- PUBLIC HELPERS --


routeToPieces : Route -> List String
routeToPieces page =
    case page of
        Overview ->
            []

        Commits branchName ->
            [ "commits", Branch.nameToString branchName ]

        Commit hash ->
            [ "commit", Commit.hashToString hash ]

        Task hash name ->
            [ "commit", Commit.hashToString hash, "tasks", ProjectTask.nameToString name ]

        Settings ->
            [ "settings" ]
