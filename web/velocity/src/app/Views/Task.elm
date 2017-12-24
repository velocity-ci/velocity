module Views.Task exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.Task as ProjectTask exposing (BuildStep, RunStep, CloneStep, Step(..), Parameter(..))


viewStepList : List Step -> Maybe Step -> List (Html msg)
viewStepList steps toggledStep =
    let
        stepView i step =
            let
                stepNum =
                    i + 1

                runStep =
                    viewRunStep stepNum

                buildStep =
                    viewBuildStep stepNum

                cloneStep =
                    viewCloneStep stepNum
            in
                case ( step, toggledStep ) of
                    ( Run run, Just (Run toggled) ) ->
                        runStep run (run == toggled)

                    ( Build build, Just (Build toggled) ) ->
                        buildStep build (build == toggled)

                    ( Clone clone, Just (Clone toggled) ) ->
                        cloneStep clone (clone == toggled)

                    ( Run run, _ ) ->
                        runStep run False

                    ( Build build, _ ) ->
                        buildStep build False

                    ( Clone clone, _ ) ->
                        cloneStep clone False
    in
        List.indexedMap stepView steps


viewCloneStep : Int -> CloneStep -> Bool -> Html msg
viewCloneStep i cloneStep toggled =
    let
        title =
            toString i ++ ". Clone" ++ cloneStep.description
    in
        viewStepCollapse (Clone cloneStep) title toggled <|
            []


viewBuildStep : Int -> BuildStep -> Bool -> Html msg
viewBuildStep i buildStep toggled =
    let
        tagList =
            List.map (\t -> li [] [ text t ]) buildStep.tags
                |> ul []

        rightDl =
            dl []
                [ dt [] [ text "Tags" ]
                , dd [] [ tagList ]
                ]

        leftDl =
            dl []
                [ dt [] [ text "Context" ]
                , dd [] [ text buildStep.context ]
                , dt [] [ text "Dockerfile" ]
                , dd [] [ text buildStep.dockerfile ]
                ]

        title =
            toString i ++ ". " ++ buildStep.description
    in
        viewStepCollapse (ProjectTask.Build buildStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ rightDl ]
                ]
            ]


viewRunStep : Int -> RunStep -> Bool -> Html msg
viewRunStep i runStep toggled =
    let
        command =
            String.join " " runStep.command

        envTable =
            table [ class "table" ]
                [ tbody []
                    (List.map
                        (\( k, v ) ->
                            tr []
                                [ th [] [ text k ]
                                , td [] [ text v ]
                                ]
                        )
                        runStep.environment
                    )
                ]

        ignoreExitCode =
            runStep.ignoreExitCode
                |> toString
                |> String.toLower

        leftDl =
            dl []
                [ dt [] [ text "Ignore exit code" ]
                , dd [] [ text ignoreExitCode ]
                , dt [] [ text "Image" ]
                , dd [] [ text runStep.image ]
                , dt [] [ text "Mount point" ]
                , dd [] [ text runStep.mountPoint ]
                , dt [] [ text "Working dir" ]
                , dd [] [ text runStep.workingDir ]
                , dt [] [ text "Command" ]
                , dd [] [ text command ]
                ]

        title =
            toString i ++ ". " ++ runStep.description
    in
        viewStepCollapse (Run runStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ envTable ]
                ]
            ]


viewStepCollapse : Step -> String -> Bool -> List (Html msg) -> Html msg
viewStepCollapse step title toggled contents =
    let
        caretClassList =
            [ ( "fa-caret-square-o-down", toggled )
            , ( "fa-caret-square-o-up", not toggled )
            ]
    in
        div [ class "card" ]
            [ div [ class "card-header collapse-header d-flex justify-content-between align-items-center" ]
                [ h5 [ class "mb-0" ] [ text title ]
                , button
                    [ type_ "button"
                    , class "btn"
                    ]
                    [ i [ class "fa", classList caretClassList ] []
                    ]
                ]
            , div
                [ class "collapse"
                , classList [ ( "show", toggled ) ]
                ]
                [ div [ class "card-body" ] contents
                ]
            ]
