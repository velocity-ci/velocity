module Page.Project.Route exposing (Route(..), routeToPieces, route, default)

import UrlParser as Url exposing (parseHash, s, (</>), (<?>), string, stringParam, intParam, oneOf, Parser)
import Data.Commit as Commit
import Data.Task as ProjectTask
import Data.Branch as Branch
import Util exposing ((=>))
import Page.Project.Commit.Route as CommitRoute


type Route
    = Overview
    | Commits (Maybe Branch.Name) (Maybe Int)
    | Commit Commit.Hash CommitRoute.Route
      --    | Task Commit.Hash ProjectTask.Name
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
        , Url.map Commit (s "commit" </> Commit.hashParser </> CommitRoute.route)
          --        , Url.map Task (s "commit" </> Commit.hashParser </> CommitRoute.route)
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

        Commit hash child ->
            let
                ( subPath, subQuery ) =
                    CommitRoute.routeToPieces child
            in
                [ "commit", Commit.hashToString hash ] ++ subPath => subQuery

        --
        --        Task hash name ->
        --            [ "commit", Commit.hashToString hash, "tasks", ProjectTask.nameToString name ] => []
        Settings ->
            [ "settings" ] => []
