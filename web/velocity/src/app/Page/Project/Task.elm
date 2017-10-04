module Page.Project.Task exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask exposing (Step(..), BuildStep, RunStep, Parameter)
import Request.Project
import Http
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Views.Page as Page
import Task exposing (Task)
import Util exposing ((=>))
import Page.Project.Route as ProjectRoute
import Route exposing (Route)
import Page.Project.Commit as Commit
import Views.Form as Form
import Validate exposing (..)
import Page.Helpers exposing (validClasses)


-- MODEL --


type alias Model =
    { commit : Commit
    , task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : List FormField
    , errors : List Error
    }


type alias FormField =
    { value : String
    , dirty : Bool
    , field : String
    }


init : Session -> Project.Id -> Commit.Hash -> ProjectTask.Name -> Task PageLoadError Model
init session id hash name =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommit =
            maybeAuthToken
                |> Request.Project.commit id hash
                |> Http.toTask

        loadTask =
            maybeAuthToken
                |> Request.Project.commitTask id hash name
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."

        initialModel commit task =
            let
                toggledStep =
                    Nothing

                form =
                    List.map newField task.parameters

                errors =
                    List.concatMap validator form
            in
                { commit = commit
                , task = task
                , toggledStep = toggledStep
                , form = form
                , errors = errors
                }
    in
        Task.map2 initialModel loadCommit loadTask
            |> Task.mapError handleLoadError


newField : Parameter -> FormField
newField parameter =
    let
        value =
            Maybe.withDefault "" parameter.default

        dirty =
            String.length value > 0
    in
        FormField value dirty parameter.name



-- VIEW --


view : Model -> Html Msg
view model =
    let
        task =
            model.task

        stepList =
            viewStepList task.steps model.toggledStep

        buildForm =
            div [ class "card" ]
                [ div [ class "card-body" ] <|
                    viewBuildForm (ProjectTask.nameToString task.name) model.form model.errors
                ]
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12 default-margin-bottom" ] [ buildForm ]
            , div [ class "col-sm-12 col-md-12 col-lg-12" ] stepList
            ]


viewBuildForm : String -> List FormField -> List Error -> List (Html Msg)
viewBuildForm taskName fields errors =
    let
        fieldInput field =
            Form.input
                { name = field.field
                , label = field.field
                , help = Nothing
                , errors = List.filter (\e -> Tuple.first e == field.field) errors
                }
                [ attribute "required" ""
                , value field.value
                , onInput (OnInput field)
                , classList (validClasses errors field)
                ]
                []
    in
        [ h4 [] [ text taskName ]
        , Html.form [ attribute "novalidate" "" ] <|
            List.map fieldInput fields
        , button
            [ class "btn btn-primary"
            , type_ "submit"
            , disabled <| not (List.isEmpty errors)
            ]
            [ text "Start task" ]
        ]


viewStepList : List Step -> Maybe Step -> List (Html Msg)
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
            in
                case ( step, toggledStep ) of
                    ( Run run, Just (Run toggled) ) ->
                        runStep run (run == toggled)

                    ( Build build, Just (Build toggled) ) ->
                        buildStep build (build == toggled)

                    ( Run run, _ ) ->
                        runStep run False

                    ( Build build, _ ) ->
                        buildStep build False
    in
        List.indexedMap stepView steps


viewBuildStep : Int -> BuildStep -> Bool -> Html Msg
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
        viewStepCollapse (Build buildStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ rightDl ]
                ]
            ]


viewRunStep : Int -> RunStep -> Bool -> Html Msg
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


viewStepCollapse : Step -> String -> Bool -> List (Html Msg) -> Html Msg
viewStepCollapse step title toggled contents =
    let
        msg =
            if toggled then
                ToggleStep Nothing
            else
                ToggleStep (Just step)

        caretClassList =
            [ ( "fa-caret-square-o-down", toggled )
            , ( "fa-caret-square-o-up", not toggled )
            ]
    in
        div [ class "card" ]
            [ div [ class "card-header collapse-header d-flex justify-content-between align-items-center", onClick msg ]
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


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route, String )
breadcrumb project commit task =
    List.concat
        [ Commit.breadcrumb project commit
        , [ ( Route.Project (ProjectRoute.Task commit.hash task.name) project.id, ProjectTask.nameToString task.name ) ]
        ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput FormField String


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        ToggleStep maybeStep ->
            { model | toggledStep = maybeStep }
                => Cmd.none

        OnInput field value ->
            let
                form =
                    List.map
                        (\f ->
                            if f == field then
                                { field
                                    | value = value
                                    , dirty = True
                                }
                            else
                                field
                        )
                        model.form

                errors =
                    List.concatMap validator form
            in
                { model
                    | form = form
                    , errors = errors
                }
                    => Cmd.none



-- VALIDATION --


type alias Error =
    ( String, String )


validator : Validator Error FormField
validator =
    Validate.all
        [ (\{ field, value } ->
            (value |> ifBlank (field => "Field cannot be blank"))
          )
        ]
