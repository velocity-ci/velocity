module Views.Build exposing (buildCardClassList, buildStepBorderColourClassList, genericToast, headerBackgroundColourClassList, streamBadgeClass, toast, viewBuildHistoryTable, viewBuildHistoryTableHeaderRow, viewBuildHistoryTableRow, viewBuildStatusIcon, viewBuildStatusIconClasses, viewBuildStepBorderClass, viewBuildStepStatusIcon, viewBuildTextClass)

import Ansi.Log
import Data.Build as Build exposing (Build, Status(..))
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStreamOutput)
import Data.Commit as Commit exposing (Commit)
import Data.Event as Event exposing (Event(..))
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Html exposing (..)
import Html.Attributes exposing (..)
import Page.Helpers exposing (formatDateTime)
import Page.Project.Commit.Route as CommitRoute
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Util exposing ((=>))
import Views.Helpers exposing (onClickPage)
import Views.Toast as Toast


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
    tr [ class "d-flex" ]
        [ th [ class "pl-0 border-0 col-6" ] [ text "Task" ]
        , th [ class "pl-0 border-0 col-2" ] [ text "Commit" ]
        , th [ class "pl-0 border-0 col-3" ] [ text "Created" ]
        , th [ class "pl-0 border-0 col-1" ] []
        ]


viewBuildHistoryTableRow : Project -> (String -> msg) -> Build -> Html msg
viewBuildHistoryTableRow project newUrlMsg build =
    let
        colourClassList =
            [ viewBuildTextClass build => True ]

        commitTaskRoute =
            CommitRoute.Task build.task.name (Just build.id)
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
        tr [ classList colourClassList, class "d-flex" ]
            [ td [ class "px-0 col-6" ] [ buildLink taskName commitTaskRoute ]
            , td [ class "px-0 col-2" ] [ buildLink truncatedHash commitRoute ]
            , td [ class "px-0 col-3" ] [ buildLink createdAt commitTaskRoute ]
            , td [ class "px-0 col-1 text-right" ] [ viewBuildStatusIcon build ]
            ]


toast : Event Build -> Html msg
toast event =
    case event of
        Event.Created build ->
            genericToast "bg-light text-dark" "Build created" build

        Event.Completed build ->
            case build.status of
                Failed ->
                    genericToast "bg-danger text-white" "Build failed" build

                Success ->
                    genericToast "bg-success text-white" "Build complete" build

                _ ->
                    genericToast "bg-light text-dark" "" build


genericToast : String -> String -> Build -> Html msg
genericToast variantClass message build =
    let
        title =
            ProjectTask.nameToString build.task.name
    in
        Toast.genericToast variantClass title message


viewBuildStatusIcon : Build -> Html msg
viewBuildStatusIcon build =
    i [ class (viewBuildStatusIconClasses build) ] []


viewBuildStatusIconClasses : Build -> String
viewBuildStatusIconClasses { status } =
    case status of
        Build.Waiting ->
            "fa fa-clock-o"

        Build.Running ->
            "fa fa-cog fa-spin"

        Build.Success ->
            "fa fa-check"

        Build.Failed ->
            "fa fa-times"


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


viewBuildStepBorderClass : BuildStep -> String
viewBuildStepBorderClass buildStep =
    case buildStep.status of
        BuildStep.Waiting ->
            "border-secondary"

        BuildStep.Running ->
            "border-primary"

        BuildStep.Success ->
            "border-success"

        BuildStep.Failed ->
            "border-danger"


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
            []

        BuildStep.Running ->
            []

        BuildStep.Success ->
            []

        BuildStep.Failed ->
            []


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
