import { Injectable } from '@angular/core';
import 'rxjs/add/operator/toPromise';
import { Router } from '@angular/router';
import {reject, resolve} from 'q';
import { APIService } from "../api.service";

@Injectable()
export class AuthService {

  constructor(
    private apiService: APIService,
    private router: Router,

  ) {}

  public authenticate(user): Promise<any> {
    return this.apiService.post(
        '/v1/auth',
        user
      ).then(
        authResponse => {
          this.apiService.setAuthorization(authResponse.json().data.authToken);
          this.router.navigate(['']);
        } 
      );
  }
}

class AuthRes {
    username: string;
    authToken: string;
    exp: string;
}
