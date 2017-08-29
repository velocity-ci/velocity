import { Component, OnInit } from '@angular/core';
import { Validators, FormControl, FormGroup } from "@angular/forms";
import { User } from "../models/user.model";
import { AuthService } from "./authentication.service";

@Component({
    selector: 'app-sign-in',
    templateUrl: './sign-in.component.html'
})
export class SignInComponent implements OnInit {

    loginForm: FormGroup;
    loginUser = new User();

    submitting = false;

    constructor(private authService: AuthService){}

    ngOnInit(): void {
        this.loginForm = new FormGroup({
            'username': new FormControl(this.loginUser.username, [
                Validators.required,
                Validators.minLength(3)
            ]),
            'password': new FormControl(this.loginUser.password, [
                Validators.required,
                Validators.minLength(8)
            ])
        })
    }

    submit() {
        this.submitting = true;
        this.authService.authenticate(this.loginUser).then(() => {

        }).catch((response) => {
            console.log(response);
            switch(response.status) {
                case 401:
                    this.loginForm.controls.username.setErrors({
                        'username': "invalid"
                    });
                    this.loginForm.controls.password.setErrors({
                        'password': "invalid"
                    });
            }
            this.submitting = false;
        });
    }
}