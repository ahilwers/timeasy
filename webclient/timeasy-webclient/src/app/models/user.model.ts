export interface User {
  id?: string;
  name: string;
  email: string;
  username: string;
  firstName?: string;
  lastName?: string;
  language?: string;
  attributes?: { [key: string]: string[] };
}
