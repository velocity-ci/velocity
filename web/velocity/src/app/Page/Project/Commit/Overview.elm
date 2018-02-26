module Page.Project.Commit.Overview exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Data.Build as Build exposing (Build)
import Page.Helpers exposing (formatDateTime)
import Navigation
import Util exposing ((=>))
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Helpers exposing (onClickPage)
import Views.Build exposing (viewBuildStatusIcon, viewBuildTextClass)


-- MODEL --


type alias Model =
    {}


initialModel : Model
initialModel =
    {}



-- VIEW --


view : Project -> Commit -> List ProjectTask.Task -> List Build -> Html Msg
view project commit tasks builds =
    div []
        [ viewCommitDetails commit
          --        , viewBuildTable project commit builds
        , viewTaskList project commit tasks builds
        ]


viewCommitDetails : Commit -> Html Msg
viewCommitDetails commit =
    let
        viewCommitDetailsIcon_ =
            viewCommitDetailsIcon commit
    in
        div [ class "card mt-3 bg-light" ]
            [ div [ class "card-body d-flex justify-content-between" ]
                [ ul [ class "list-unstyled mb-0" ]
                    [ viewCommitDetailsIcon_ "fa-comment-o" .message
                    , viewCommitDetailsIcon_ "fa-file-code-o" (.hash >> Commit.hashToString)
                    , viewCommitDetailsIcon_ "fa-user" .author
                    , viewCommitDetailsIcon_ "fa-calendar" (.date >> formatDateTime)
                    ]
                ]
            ]


viewCommitDetailsIcon : Commit -> String -> (Commit -> String) -> Html Msg
viewCommitDetailsIcon commit iconClass fn =
    li []
        [ i
            [ attribute "aria-hidden" "true"
            , classList
                [ ( "fa", True )
                , ( iconClass, True )
                ]
            ]
            []
        , " " ++ fn commit |> text
        ]


viewTaskList : Project -> Commit -> List ProjectTask.Task -> List Build -> Html Msg
viewTaskList project commit tasks builds =
    let
        taskList =
            List.map (viewTaskListItem project commit builds) tasks
                |> div [ class "list-group list-group-flush" ]
    in
        div [ class "card mt-3" ]
            [ h5 [ class "card-header" ] [ text "Tasks" ]
            , taskList
            ]



--
--viewBuildTable : Project -> Commit -> List Build -> Html Msg
--viewBuildTable project commit builds =
--    let
--        header =
--            thead []
--                [ tr []
--                    [ th [ scope "col" ] [ text "#" ]
--                    , th [ scope "col" ] [ text "Task" ]
--                    , th [ scope "col" ] [ text "Status" ]
--                    ]
--                ]
--    in
--        table [ class "table table-bordered mt-3" ]
--            [ header
--            , tbody [] (List.map (viewBuildTableRow project commit) builds)
--            ]
--
--
--viewBuildTableRow : Project -> Commit -> Build -> Html Msg
--viewBuildTableRow project commit build =
--    let
--        route =
--            CommitRoute.Build build.id
--                |> ProjectRoute.Commit commit.hash
--                |> Route.Project project.id
--    in
--        tr []
--            [ td []
--                [ a
--                    [ Route.href route
--                    , onClickPage NewUrl route
--                    ]
--                    [ text <| Build.idToString build.id ]
--                ]
--            , td [] [ text <| ProjectTask.nameToString build.task ]
--            , td [] []
--            ]


taskBuilds : ProjectTask.Task -> List Build -> List Build
taskBuilds task builds =
    builds
        |> List.filter (\b -> ProjectTask.idEquals task.id b.task.id)


maybeBuildFromTask : ProjectTask.Task -> List Build -> Maybe Build
maybeBuildFromTask task builds =
    taskBuilds task builds
        |> List.head


viewTaskListItem : Project -> Commit -> List Build -> ProjectTask.Task -> Html Msg
viewTaskListItem project commit builds task =
    let
        buildNum =
            List.length (taskBuilds task builds)

        routeTabParam =
            if buildNum > 0 then
                Just ("build-" ++ (toString buildNum))
            else
                Nothing

        route =
            CommitRoute.Task task.name routeTabParam
                |> ProjectRoute.Commit commit.hash
                |> Route.Project project.slug

        maybeBuild =
            maybeBuildFromTask task builds

        icon =
            maybeBuild
                |> Maybe.map viewBuildStatusIcon
                |> Maybe.withDefault (text "")

        textClass =
            maybeBuild
                |> Maybe.map viewBuildTextClass
                |> Maybe.withDefault ""
    in
        a
            [ class (textClass ++ " list-group-item list-group-item-action flex-column align-items-center justify-content-between")
            , Route.href route
            , onClickPage NewUrl route
            ]
            [ div [ class "" ] [ h5 [ class "mb-1" ] [ text (ProjectTask.nameToString task.name) ] ]
            , p [ class "mb-1" ] [ text task.description ]
            , icon
            ]



-- UPDATE --


type Msg
    = NewUrl String


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        NewUrl url ->
            model => Navigation.newUrl url
