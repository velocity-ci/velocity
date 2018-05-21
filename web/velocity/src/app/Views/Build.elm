module Views.Build exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Build as Build exposing (Build)
import Data.Project as Project exposing (Project)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStreamOutput)
import Data.Commit as Commit exposing (Commit)
import Data.Task as ProjectTask
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Page.Helpers exposing (formatDateTime)
import Ansi.Log
import Util exposing ((=>))
import Views.Helpers exposing (onClickPage)


viewBuildHistoryTable : Project -> List Build -> (String -> msg) -> Html msg
viewBuildHistoryTable project builds newUrlMsg =
    div [ class "col-md-12 px-0 mx-0" ]
        [ div []
            [ table [ class "table mb-0 " ]
                [ thead [] [ viewBuildHistoryTableHeaderRow ]
                , tbody [] (List.map (viewBuildHistoryTableRow project newUrlMsg) builds)
                ]
            ]
        ]


viewBuildHistoryTableHeaderRow : Html msg
viewBuildHistoryTableHeaderRow =
    tr []
        [ th [ class "pl-0 border-0" ] [ text "Task" ]
        , th [ class "pl-0 border-0" ] [ text "Commit" ]
        , th [ class "pl-0 border-0" ] [ text "Created" ]
        , th [ class "pl-0 border-0" ] []
        ]


viewBuildHistoryTableRow : Project -> (String -> msg) -> Build -> Html msg
viewBuildHistoryTableRow project newUrlMsg build =
    let
        colourClassList =
            [ viewBuildTextClass build => True ]

        commitTaskRoute =
            CommitRoute.Task build.task.name Nothing
                |> ProjectRoute.Commit build.task.commit.hash
                |> Route.Project project.slug

        commitRoute =
            CommitRoute.Overview
                |> ProjectRoute.Commit build.task.commit.hash
                |> Route.Project project.slug

        task =
            build.task

        taskName =
            ProjectTask.nameToString task.name

        createdAt =
            formatDateTime build.createdAt

        truncatedHash =
            Commit.truncateHash task.commit.hash

        buildLink content route =
            a
                [ Route.href route
                , onClickPage newUrlMsg route
                , classList colourClassList
                ]
                [ text content ]
    in
        tr [ classList colourClassList ]
            [ td [ class "px-0" ] [ buildLink taskName commitTaskRoute ]
            , td [ class "px-0" ] [ buildLink truncatedHash commitRoute ]
            , td [ class "px-0" ] [ buildLink createdAt commitTaskRoute ]
            , td [ class "px-0 text-right" ] [ viewBuildStatusIcon build ]
            ]


viewBuildStatusIcon : Build -> Html msg
viewBuildStatusIcon build =
    case build.status of
        Build.Waiting ->
            i [ class "fa fa-clock-o" ] []

        Build.Running ->
            i [ class "fa fa-cog fa-spin" ] []

        Build.Success ->
            i [ class "fa fa-check" ] []

        Build.Failed ->
            i [ class "fa fa-times" ] []


viewBuildTextClass : Build -> String
viewBuildTextClass build =
    case build.status of
        Build.Waiting ->
            "text-secondary"

        Build.Running ->
            "text-primary"

        Build.Success ->
            "text-success"

        Build.Failed ->
            "text-danger"


viewBuildStepStatusIcon : BuildStep -> Html msg
viewBuildStepStatusIcon buildStep =
    case buildStep.status of
        BuildStep.Waiting ->
            i [ class "fa fa-clock-o" ] []

        BuildStep.Running ->
            i [ class "fa fa-cog fa-spin" ] []

        BuildStep.Success ->
            i [ class "fa fa-check" ] []

        BuildStep.Failed ->
            i [ class "fa fa-times" ] []


streamBadgeClass : Int -> String
streamBadgeClass index =
    case index of
        0 ->
            "badge-primary"

        1 ->
            "badge-secondary"

        2 ->
            "badge-success"

        3 ->
            "badge-danger"

        4 ->
            "badge-warning"

        5 ->
            "badge-info"

        _ ->
            "badge-dark"


headerBackgroundColourClassList : BuildStep -> List ( String, Bool )
headerBackgroundColourClassList { status } =
    case status of
        BuildStep.Waiting ->
            []

        BuildStep.Running ->
            []

        BuildStep.Success ->
            [ "text-success" => True
            , "bg-transparent" => True
            ]

        BuildStep.Failed ->
            [ "bg-transparent" => True
            , "text-danger" => True
            ]


buildStepBorderColourClassList : BuildStep -> List ( String, Bool )
buildStepBorderColourClassList { status } =
    case status of
        BuildStep.Waiting ->
            [ "border" => True
            ]

        BuildStep.Running ->
            [ "border" => True
            , "border-primary" => True
            ]

        BuildStep.Success ->
            [ "border" => True
            ]

        BuildStep.Failed ->
            [ "border" => True
            ]


buildCardClassList : Build -> List ( String, Bool )
buildCardClassList { status } =
    case status of
        Build.Success ->
            [ "text-success" => True
            ]

        Build.Failed ->
            [ "text-danger" => True
            ]

        _ ->
            []


viewCardTitle : ProjectTask.Step -> String
viewCardTitle taskStep =
    case taskStep of
        ProjectTask.Build _ ->
            "Build"

        ProjectTask.Run _ ->
            "Run"

        ProjectTask.Clone _ ->
            "Clone"

        ProjectTask.Compose _ ->
            "Compose"

        ProjectTask.Push _ ->
            "Push"
