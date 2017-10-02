module Page.Project.Route exposing (Route(..), routeToPieces)

import Data.Commit as Commit
import Data.Task as ProjectTask


type Route
    = Overview
    | Commits
    | Commit Commit.Hash
    | Task Commit.Hash ProjectTask.Name
    | Settings



-- PUBLIC HELPERS --


routeToPieces : Route -> List String
routeToPieces page =
    case page of
        Overview ->
            []

        Commits ->
            [ "commits" ]

        Commit hash ->
            [ "commits", Commit.hashToString hash ]

        Task hash name ->
            [ "commits", Commit.hashToString hash, "tasks", ProjectTask.nameToString name ]

        Settings ->
            [ "settings" ]
