module Page.Project.Route exposing (Route(..), routeToPieces, route, default)

import UrlParser as Url exposing (parseHash, s, (</>), (<?>), string, stringParam, intParam, oneOf, Parser)
import Data.Commit as Commit
import Data.Task as ProjectTask
import Data.Branch as Branch
import Util exposing ((=>))


type Route
    = Overview
    | Commits (Maybe Branch.Name) (Maybe Int)
    | Commit Commit.Hash
    | Task Commit.Hash ProjectTask.Name
    | Settings


default : Route
default =
    Overview


route : Parser (Route -> b) b
route =
    oneOf
        [ Url.map Overview (s "overview")
        , Url.map Settings (s "settings")
        , Url.map Commits (s "commits" </> Branch.nameParser <?> intParam "page")
        , Url.map Commit (s "commit" </> Commit.hashParser)
        , Url.map Task (s "commit" </> Commit.hashParser </> s "tasks" </> ProjectTask.nameParser)
        ]



-- PUBLIC HELPERS --


routeToPieces : Route -> ( List String, List ( String, String ) )
routeToPieces page =
    case page of
        Overview ->
            [] => []

        Commits branchName maybePageNumber ->
            let
                queryParams =
                    case maybePageNumber of
                        Just 1 ->
                            []

                        Just p ->
                            [ ( "page", toString p ) ]

                        _ ->
                            []
            in
                [ "commits", Branch.nameToString branchName ] => queryParams

        Commit hash ->
            [ "commit", Commit.hashToString hash ] => []

        Task hash name ->
            [ "commit", Commit.hashToString hash, "tasks", ProjectTask.nameToString name ] => []

        Settings ->
            [ "settings" ] => []
