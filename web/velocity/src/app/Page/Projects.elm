module Page.Projects exposing (..)

import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Views.Page as Page
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, onSubmit)
import Util exposing ((=>))
import Views.Form as Form
import Validate exposing (..)
import Page.Helpers exposing (ifBelowLength, validClasses, formatDateTime, sortByDatetime)


-- MODEL --


type Field
    = Form
    | Name
    | Repository
    | PrivateKey


type alias FormField =
    { value : String
    , dirty : Bool
    , field : Field
    }


type alias Model =
    { formCollapsed : Bool
    , name : FormField
    , repository : FormField
    , privateKey : FormField
    , errors : List Error
    , submitting : Bool
    , projects : List Project
    }


newField : Field -> FormField
newField field =
    FormField "" False field


init : Session -> Task PageLoadError Model
init session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProjects =
            Request.Project.list maybeAuthToken
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Projects "Projects are currently unavailable."

        initialFormErrors =
            validate
                { name = newField Name
                , repository = newField Repository
                , privateKey = newField PrivateKey
                }

        staticInitialModel =
            Model True (newField Name) (newField Repository) (newField PrivateKey) initialFormErrors False
    in
        Task.map staticInitialModel loadProjects
            |> Task.mapError handleLoadError



-- VIEW --


view : Session -> Model -> Html Msg
view session model =
    div []
        [ viewProjectFormContainer model
        , viewProjectList model.projects
        ]


viewProjectFormContainer : Model -> Html Msg
viewProjectFormContainer model =
    let
        toggleClassList =
            [ ( "fa-plus", model.formCollapsed )
            , ( "fa-minus", not model.formCollapsed )
            ]
    in
        div [ class "row" ]
            [ div [ class "col-12" ]
                [ div [ class "card" ]
                    [ h4 [ class "card-header" ]
                        [ text "Create Project"
                        , button
                            [ type_ "button"
                            , class "btn btn-primary btn-sm float-right"
                            , onClick (SetFormCollapsed <| not model.formCollapsed)
                            ]
                            [ i [ class "fa", classList toggleClassList ] [] ]
                        ]
                    , Util.viewIf (not model.formCollapsed) <| viewProjectForm model
                    ]
                ]
            ]


viewProjectForm : Model -> Html Msg
viewProjectForm model =
    let
        inputClassList =
            validClasses <| model.errors
    in
        div [ class "card-body" ]
            [ Html.form [ attribute "novalidate" "", onSubmit SubmitForm ]
                [ Form.input
                    "name"
                    "Name"
                    [ placeholder "Name"
                    , attribute "required" ""
                    , value model.name.value
                    , onInput SetName
                    , classList <| inputClassList model.name
                    ]
                    []
                , Form.input
                    "repository"
                    "Repository address"
                    [ placeholder "Repository"
                    , attribute "required" ""
                    , value model.repository.value
                    , onInput SetRepository
                    , classList <| inputClassList model.repository
                    ]
                    []
                , Form.textarea
                    "key"
                    "Private key"
                    [ placeholder "Private key"
                    , attribute "required" ""
                    , rows 3
                    , value model.privateKey.value
                    , onInput SetPrivateKey
                    , classList <| inputClassList model.privateKey
                    ]
                    []
                , button
                    [ class "btn btn-primary"
                    , type_ "submit"
                    , disabled ((not <| List.isEmpty model.errors) && (not <| model.submitting))
                    ]
                    [ text "Submit" ]
                ]
            ]


viewProjectList : List Project -> Html Msg
viewProjectList projects =
    let
        projectAmount =
            projects
                |> List.length
                |> toString

        latestProjects =
            sortByDatetime .updatedAt projects
    in
        div [ class "row", style [ ( "margin-top", "3em" ) ] ]
            [ div [ class "col-12" ]
                [ div [ class "card" ]
                    [ h4 [ class "card-header" ] [ text ("Projects (" ++ projectAmount ++ ")") ]
                    , ul [ class "list-group" ] (List.map viewProjectListItem latestProjects)
                    ]
                ]
            ]


viewProjectListItem : Project -> Html Msg
viewProjectListItem project =
    li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h5 [ class "mb-1" ] [ a [ href "#" ] [ text project.name ] ]
            , small [] [ text (formatDateTime project.updatedAt) ]
            ]
        , small []
            [ text project.repository ]
        ]



-- UPDATE --


type Msg
    = SubmitForm
    | SetFormCollapsed Bool
    | SetName String
    | SetRepository String
    | SetPrivateKey String
    | ProjectCreated (Result Http.Error Project)


updateInput : Field -> String -> FormField
updateInput field value =
    FormField value True field


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    case msg of
        SubmitForm ->
            case validate model of
                [] ->
                    let
                        submitValues =
                            { name = model.name.value
                            , repository = model.repository.value
                            , privateKey = model.privateKey.value
                            }

                        cmdFromAuth authToken =
                            authToken
                                |> Request.Project.create submitValues
                                |> Http.send ProjectCreated

                        cmd =
                            session
                                |> Session.attempt "create project" cmdFromAuth
                                |> Tuple.second
                    in
                        { model
                            | submitting = True
                            , name = newField Name
                            , repository = newField Repository
                            , privateKey = newField PrivateKey
                            , formCollapsed = True
                        }
                            => cmd

                errors ->
                    { model | errors = errors }
                        => Cmd.none

        SetFormCollapsed state ->
            { model | formCollapsed = state }
                => Cmd.none

        SetName name ->
            let
                newModel =
                    { model | name = name |> (updateInput Name) }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none

        SetRepository repository ->
            let
                newModel =
                    { model | repository = repository |> (updateInput Repository) }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none

        SetPrivateKey privateKey ->
            let
                newModel =
                    { model | privateKey = privateKey |> (updateInput PrivateKey) }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none

        ProjectCreated (Err err) ->
            { model | submitting = False } => Cmd.none

        ProjectCreated (Ok project) ->
            { model
                | projects = project :: model.projects
                , submitting = False
            }
                => Cmd.none



-- VALIDATION --


type alias Error =
    ( Field, String )


validate :
    Validator ( Field, String )
        { d
            | name : { a | value : String }
            , privateKey : { b | value : String }
            , repository : { c | value : String }
        }
validate =
    Validate.all
        [ (.name >> .value) >> ifBlank (Name => "project name can't be blank.")
        , (.repository >> .value) >> ifBlank (Repository => "project repository can't be blank.")
        , (.privateKey >> .value) >> ifBlank (PrivateKey => "private key can't be blank.")
        , (.name >> .value) >> (ifBelowLength 3) (Name => "name must be over 2 characters.")
        , (.repository >> .value) >> (ifBelowLength 8) (Repository => "repository must be over 7 characters.")
        , (.privateKey >> .value) >> (ifBelowLength 8) (PrivateKey => "private key must be over 7 characters.")
        ]
