module Page.Project.Commit.Route exposing (Route(..), routeToPieces, route)

import UrlParser as Url exposing (parseHash, s, (</>), (<?>), string, stringParam, intParam, oneOf, Parser)
import Data.Commit as Commit
import Data.Task as ProjectTask
import Data.Build as Build
import Util exposing ((=>))


type Route
    = Overview
    | Task ProjectTask.Name
    | Build Build.Id


route : Parser (Route -> b) b
route =
    oneOf
        [ Url.map Overview (s "overview")
        , Url.map Task (s "tasks" </> ProjectTask.nameParser)
        , Url.map Build (s "builds" </> Build.idParser)
        ]



-- PUBLIC HELPERS --


routeToPieces : Route -> ( List String, List ( String, String ) )
routeToPieces page =
    case page of
        Overview ->
            [ "overview" ] => []

        Task name ->
            [ "tasks", ProjectTask.nameToString name ] => []

        Build id ->
            [ "builds", Build.idToString id ] => []
