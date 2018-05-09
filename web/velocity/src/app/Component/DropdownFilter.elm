module Component.DropdownFilter exposing (Context, Config, DropdownState, initialDropdownState, view, subscriptions)

{- A stateless ItemFilter component. -}
-- EXTERNAL --

import Html exposing (..)
import Html.Events exposing (onWithOptions, onInput)
import Html.Attributes exposing (..)
import Bootstrap.Dropdown as Dropdown
import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Json.Decode as Decode


-- MODEL --


type alias Context a =
    { items : List a
    , dropdownState : Dropdown.State
    , filterTerm : String
    , selectedItem : Maybe a
    }


type alias Config msg item =
    { dropdownMsg : Dropdown.State -> msg
    , termMsg : String -> msg
    , noOpMsg : msg
    , selectItemMsg : Maybe item -> msg
    , labelFn : item -> String
    }


type alias DropdownState =
    Dropdown.State


initialDropdownState : DropdownState
initialDropdownState =
    Dropdown.initialState



-- SUBSCRIPTIONS --


subscriptions : Config msg a -> Context a -> Sub msg
subscriptions { dropdownMsg } { dropdownState } =
    Dropdown.subscriptions dropdownState dropdownMsg



-- VIEW --


view : Config msg a -> Context a -> Html msg
view config context =
    Dropdown.dropdown
        context.dropdownState
        { options =
            [ Dropdown.menuAttrs
                [ onClick (config.noOpMsg)
                , class "item-filter-dropdown"
                ]
            ]
        , toggleMsg = config.dropdownMsg
        , toggleButton = toggleButton context config
        , items = viewDropdownItems config context
        }


toggleButton : Context a -> Config msg a -> Dropdown.DropdownToggle msg
toggleButton { selectedItem } { labelFn } =
    let
        toggleText =
            case selectedItem of
                Just b ->
                    labelFn b

                Nothing ->
                    ""
    in
        Dropdown.toggle
            [ Button.outlineSecondary ]
            [ i [ class "fa fa-code-fork" ] []
            , text (" " ++ toggleText)
            ]


viewDropdownItems : Config msg a -> Context a -> List (Dropdown.DropdownItem msg)
viewDropdownItems config context =
    let
        filterForm =
            Dropdown.customItem (viewForm config.termMsg context)

        items =
            viewItems config context

        noItemSelectedButton =
            noItemSelected config context
    in
        filterForm :: Dropdown.divider :: noItemSelectedButton :: Dropdown.divider :: items


noItemSelected : Config msg a -> Context a -> Dropdown.DropdownItem msg
noItemSelected config { selectedItem } =
    viewItem config selectedItem Nothing


viewItems : Config msg a -> Context a -> List (Dropdown.DropdownItem msg)
viewItems config { items, filterTerm, selectedItem } =
    items
        |> List.filter (config.labelFn >> String.contains filterTerm)
        |> List.map (Just >> viewItem config selectedItem)


viewItem : Config msg a -> Maybe a -> Maybe a -> Dropdown.DropdownItem msg
viewItem config selectedItem maybeItem =
    let
        itemText =
            maybeItem
                |> Maybe.map config.labelFn
                |> Maybe.withDefault "all-items"
                |> text

        itemIcon =
            if selectedItem == maybeItem then
                i [ class "fa fa-check" ] []
            else
                text ""
    in
        Dropdown.buttonItem
            [ onClick (config.selectItemMsg <| maybeItem) ]
            [ itemIcon
            , itemText
            ]


viewForm : (String -> msg) -> Context a -> Html msg
viewForm msg { filterTerm } =
    Form.form [ class "px-2 py-0 filter-list-select", style [ ( "width", "400px" ) ] ]
        [ Form.group []
            [ Input.text
                [ Input.placeholder "Filter items"
                , Input.attrs [ onInput msg ]
                , Input.value filterTerm
                ]
            ]
        ]



-- helper to cancel click anywhere


onClick : msg -> Attribute msg
onClick message =
    onWithOptions
        "click"
        { stopPropagation = True
        , preventDefault = False
        }
        (Decode.succeed message)
