module Page.Project.Commit.Task exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Task exposing (Task)
import Dict exposing (Dict)
import Json.Encode as Encode
import Bootstrap.Modal as Modal
import Navigation


-- INTERNAL --

import Context exposing (Context)
import Component.BuildOutput as BuildOutput
import Component.BuildForm as BuildForm
import Component.DropdownFilter as DropdownFilter
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.BuildStream as BuildStream exposing (Id, BuildStream, BuildStreamOutput)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Util exposing ((=>))
import Views.Page as Page
import Views.Task exposing (viewStepList)
import Page.Helpers exposing (formatDateTime, sortByDatetime)


--import Views.Build exposing (viewBuildStatusIcon, viewBuildStepStatusIcon, viewBuildTextClass)

import Request.Commit
import Request.Errors
import Route


-- MODEL --


type alias Model =
    { task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : BuildForm.Context
    , formModalVisibility : Modal.Visibility
    , selected : Maybe Build.Id
    , frame : Frame
    , buildDropdownState : DropdownFilter.DropdownState
    , buildFilterTerm : String
    }


type alias FromBuild =
    Build.Id


type alias ToBuild =
    Build.Id


type Stream
    = Stream BuildStream.Id


type Frame
    = BuildFrame BuildType
    | BlankFrame


type BuildType
    = LoadedBuild Build.Id BuildOutput.Model
    | LoadingBuild (Maybe FromBuild) (Maybe ToBuild)


init : Context -> Session msg -> Project.Id -> Commit.Hash -> ProjectTask.Task -> Maybe Build.Id -> List Build -> Task PageLoadError Model
init context session id hash task selected builds =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        init =
            maybeBuildToModel context task maybeAuthToken
    in
        case selected of
            Just id ->
                id
                    |> Build.findBuild builds
                    |> init

            Nothing ->
                builds
                    |> List.reverse
                    |> List.head
                    |> init


maybeBuildToModel :
    Context
    -> ProjectTask.Task
    -> Maybe AuthToken
    -> Maybe Build
    -> Task PageLoadError Model
maybeBuildToModel context task maybeAuthToken maybeBuild =
    let
        selected =
            Maybe.map .id maybeBuild

        init =
            initialModel task selected
    in
        case maybeBuild of
            Just b ->
                BuildOutput.init context task maybeAuthToken b
                    |> Task.map (LoadedBuild b.id >> BuildFrame >> init)
                    |> Task.mapError handleLoadError

            Nothing ->
                Task.succeed (init BlankFrame)


initialModel : ProjectTask.Task -> Maybe Build.Id -> Frame -> Model
initialModel task selected frame =
    { task = task
    , toggledStep = Nothing
    , form = BuildForm.init task
    , formModalVisibility = Modal.hidden
    , selected = selected
    , frame = frame
    , buildDropdownState = DropdownFilter.initialDropdownState
    , buildFilterTerm = ""
    }


handleLoadError : a -> PageLoadError
handleLoadError _ =
    pageLoadError Page.Project "Project unavailable."



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { formModalVisibility } =
    Modal.subscriptions formModalVisibility AnimateFormModal



-- CHANNELS --


events : Model -> Dict String (List ( String, Encode.Value -> Msg ))
events model =
    case model.frame of
        BuildFrame (LoadedBuild _ buildOutputModel) ->
            BuildOutput.events buildOutputModel
                |> mapEvents BuildOutputMsg

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe CommitRoute.Route -> List String
leaveChannels model route =
    let
        isTask task =
            model.task.name == task

        --
        --        isTab tab =
        --            model.selectedTab == (stringToTab tab)
        channels =
            case model.frame of
                BuildFrame (LoadedBuild _ buildOutputModel) ->
                    BuildOutput.leaveChannels buildOutputModel

                _ ->
                    []
    in
        case route of
            Just (CommitRoute.Task task tab) ->
                if not (isTask task) then
                    channels
                else
                    []

            _ ->
                channels



--leaveChannels : Model -> List String
--leaveChannels model =
--    case model.frame of
--        BuildFrame (LoadedBuild _ buildOutputModel) ->
--            BuildOutput.leaveChannels buildOutputModel
--
--        _ ->
--            []


mapEvents :
    (b -> c)
    -> Dict comparable (List ( a1, a -> b ))
    -> Dict comparable (List ( a1, a -> c ))
mapEvents fromMsg events =
    events
        |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)


buildDropdownFilterConfig : DropdownFilter.Config Msg Build
buildDropdownFilterConfig =
    { dropdownMsg = BuildFilterDropdownMsg
    , termMsg = BuildFilterTermMsg
    , noOpMsg = NoOp_
    , selectItemMsg = always NoOp_
    , labelFn = (.createdAt >> formatDateTime)
    , icon = (strong [] [ text "Build: " ])
    , showFilter = False
    , showAllItemsItem = False
    }


buildFilterContext : Model -> List Build -> DropdownFilter.Context Build
buildFilterContext { frame, buildDropdownState, buildFilterTerm, selected } builds =
    { items = sortByDatetime .createdAt builds
    , dropdownState = buildDropdownState
    , filterTerm = buildFilterTerm
    , selectedItem = Maybe.andThen (Build.findBuild builds) selected
    }



--
-- VIEW --


view : Project -> Commit -> Model -> List Build -> Html Msg
view project commit model builds =
    let
        task =
            model.task

        stepList =
            viewStepList task.steps model.toggledStep

        dropdownContext =
            buildFilterContext model builds
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12" ]
                [ DropdownFilter.view buildDropdownFilterConfig dropdownContext
                , viewToolbar
                , viewTabFrame model builds commit
                , viewFormModal model.task model.form model.formModalVisibility
                ]
            ]


viewFormModal : ProjectTask.Task -> BuildForm.Context -> Modal.Visibility -> Html Msg
viewFormModal task form visibility =
    let
        hasFields =
            not (List.isEmpty form.fields)

        basicModal =
            Modal.config CloseFormModal
                |> Modal.withAnimation AnimateFormModal
                |> Modal.large
                |> Modal.hideOnBackdropClick True
                |> Modal.h5 []
                    [ text "Start "
                    , strong [] [ text <| ProjectTask.nameToString task.name ]
                    ]
                |> Modal.footer [] [ BuildForm.viewSubmitButton buildFormConfig form ]

        noParametersAlert =
            div [ class "alert alert-info m-0" ]
                [ i [ class "fa fa-info-circle" ] []
                , text " No parameters required"
                ]

        modal =
            if hasFields then
                Modal.body [] (BuildForm.view buildFormConfig form) basicModal
            else
                Modal.body [] [ noParametersAlert ] basicModal
    in
        Modal.view visibility modal


viewToolbar : Html Msg
viewToolbar =
    div [ class "btn-toolbar d-flex flex-row-reverse" ]
        [ button
            [ class "btn btn-primary btn-lg"
            , style [ "border-radius" => "25px" ]
            , onClick ShowOrSubmitTaskForm
            ]
            [ i [ class "fa fa-plus" ] [] ]
        ]


viewTabFrame : Model -> List Build -> Commit -> Html Msg
viewTabFrame model builds commit =
    let
        findBuild id =
            builds
                |> List.filter (\a -> a.id == id)
                |> List.head
    in
        if List.isEmpty builds then
            viewNoBuildsAlert model.task commit
        else
            case model.frame of
                BlankFrame ->
                    text ""

                BuildFrame (LoadedBuild buildId buildOutputModel) ->
                    case findBuild buildId of
                        Just build ->
                            BuildOutput.view build buildOutputModel
                                |> Html.map BuildOutputMsg

                        Nothing ->
                            text ""

                BuildFrame (LoadingBuild _ _) ->
                    text ""


viewNoBuildsAlert : ProjectTask.Task -> Commit -> Html Msg
viewNoBuildsAlert task commit =
    let
        icon =
            i [ class "fa fa-info-circle" ] []

        preTaskNameText =
            text " Task "

        taskNameText =
            strong [] [ text (ProjectTask.nameToString task.name) ]

        postTaskNameText =
            text " has not run yet for commit "

        commitShaText =
            strong [] [ text (Commit.truncateHash commit.hash) ]
    in
        div [ class "alert alert-info mt-4" ]
            [ icon
            , preTaskNameText
            , taskNameText
            , postTaskNameText
            , commitShaText
            ]


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route.Route, String )
breadcrumb project commit task =
    [ ( CommitRoute.Task task.name Nothing |> ProjectRoute.Commit commit.hash |> Route.Project project.slug
      , ProjectTask.nameToString task.name
      )
    ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput BuildForm.InputFormField String
    | OnChange BuildForm.ChoiceFormField (Maybe Int)
    | SubmitForm
    | BuildCreated (Result Request.Errors.HttpError Build)
      --    | SelectTab Tab String
    | SelectBuild (Maybe Build)
    | BuildLoaded (Result Request.Errors.HttpError (Maybe BuildType))
    | BuildOutputMsg BuildOutput.Msg
    | CloseFormModal
    | AnimateFormModal Modal.Visibility
    | ShowOrSubmitTaskForm
    | BuildFilterDropdownMsg DropdownFilter.DropdownState
    | BuildFilterTermMsg String
    | NoOp_


type ExternalMsg
    = NoOp
    | AddBuild Build
    | UpdateBuild Build


buildFormConfig : BuildForm.Config Msg
buildFormConfig =
    { submitMsg = SubmitForm
    , onChangeMsg = OnChange
    , onInputMsg = OnInput
    }


update : Context -> Project -> Commit -> List Build -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context project commit builds session msg model =
    let
        projectSlug =
            project.slug

        commitHash =
            commit.hash

        taskName =
            model.task.name

        maybeAuthToken =
            Maybe.map .token session.user
    in
        case msg of
            ShowOrSubmitTaskForm ->
                let
                    form =
                        BuildForm.init model.task
                in
                    { model
                        | formModalVisibility = Modal.shown
                        , form = form
                    }
                        => Cmd.none
                        => NoOp

            CloseFormModal ->
                { model
                    | formModalVisibility = Modal.hidden
                    , form = BuildForm.init model.task
                }
                    => Cmd.none
                    => NoOp

            AnimateFormModal visibility ->
                { model | formModalVisibility = visibility }
                    => Cmd.none
                    => NoOp

            ToggleStep maybeStep ->
                { model | toggledStep = maybeStep }
                    => Cmd.none
                    => NoOp

            OnInput field value ->
                { model | form = BuildForm.updateInput field value model.form }
                    => Cmd.none
                    => NoOp

            OnChange field maybeIndex ->
                { model | form = BuildForm.updateSelect field maybeIndex model.form }
                    => Cmd.none
                    => NoOp

            SubmitForm ->
                let
                    cmdFromAuth authToken =
                        authToken
                            |> Request.Commit.createBuild context projectSlug commitHash taskName (BuildForm.submitParams model.form)
                            |> Task.attempt BuildCreated

                    cmd =
                        session
                            |> Session.attempt "create build" cmdFromAuth
                            |> Tuple.second
                in
                    model
                        => cmd
                        => NoOp

            BuildLoaded (Ok (Just loadedBuild)) ->
                model
                    => Cmd.none
                    => NoOp

            BuildLoaded _ ->
                model
                    => Cmd.none
                    => NoOp

            BuildCreated (Ok build) ->
                let
                    route =
                        CommitRoute.Task model.task.name (Just build.id)
                            |> ProjectRoute.Commit commit.hash
                            |> Route.Project project.slug
                in
                    model
                        => Navigation.newUrl (Route.routeToString route)
                        => AddBuild build

            BuildCreated (Err _) ->
                model
                    => Cmd.none
                    => NoOp

            --            SelectTab tab url ->
            --                let
            --                    frame =
            --                        case tab of
            --                            BuildTab toBuildIndex ->
            --                                let
            --                                    toBuild =
            --                                        builds
            --                                            |> Array.fromList
            --                                            |> Array.get (toBuildIndex - 1)
            --                                            |> Maybe.map .id
            --
            --                                    fromBuild =
            --                                        case model.frame of
            --                                            BuildFrame (LoadedBuild b _) ->
            --                                                Just b
            --
            --                                            _ ->
            --                                                Nothing
            --                                in
            --                                    BuildFrame (LoadingBuild fromBuild toBuild)
            --                in
            --                    { model
            --                        | selectedTab = Just tab
            --                        , frame = frame
            --                    }
            --                        => Navigation.newUrl url
            --                        => NoOp
            BuildOutputMsg subMsg ->
                case model.frame of
                    BuildFrame (LoadedBuild id outputModel) ->
                        let
                            ( newOutputModel, newOutputCmd ) =
                                BuildOutput.update subMsg outputModel
                        in
                            { model | frame = BuildFrame (LoadedBuild id newOutputModel) }
                                => Cmd.map BuildOutputMsg newOutputCmd
                                => NoOp

                    _ ->
                        model
                            => Cmd.none
                            => NoOp

            BuildFilterDropdownMsg state ->
                { model | buildDropdownState = state }
                    => Cmd.none
                    => NoOp

            BuildFilterTermMsg term ->
                { model | buildFilterTerm = term }
                    => Cmd.none
                    => NoOp

            SelectBuild maybeBuild ->
                let
                    route =
                        CommitRoute.Task model.task.name Nothing
                            |> ProjectRoute.Commit commitHash
                            |> Route.Project project.slug
                in
                    model
                        => Route.modifyUrl route
                        => NoOp

            NoOp_ ->
                model
                    => Cmd.none
                    => NoOp
