package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PedPet/api-gateway/config"
	userSvcEndpoint "github.com/PedPet/user/pkg/endpoint"
	grpcClient "github.com/PedPet/user/pkg/grpc"
	userSvcService "github.com/PedPet/user/pkg/service"
	"google.golang.org/grpc"
)

// User endpoints
type User struct {
	RegisterEndpoint           http.HandlerFunc
	ConfirmUserEndpoint        http.HandlerFunc
	ResendConfirmationEndpoint http.HandlerFunc
	CheckUsernameEndpoint      http.HandlerFunc
	LoginEndpoint              http.HandlerFunc
	// VerifyJWTEndpoint          http.HandlerFunc
	// UserDetails                http.HandlerFunc
}

func userService(us config.Service) grpcSvcConnFunc {
	return func() (userSvcService.User, *grpc.ClientConn, error) {
		conn, err := grpc.Dial(us.Host+":"+us.Port, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		userSvc := grpcClient.NewClient(conn)

		return userSvc, conn, nil
	}
}

func makeUserEndpoints(ctx context.Context, us config.Service) User {
	svcFunc := userService(us)

	return User{
		RegisterEndpoint:           makeRegisterEndpoint(ctx, svcFunc),
		ConfirmUserEndpoint:        makeConfirmUserEndpoint(ctx, svcFunc),
		ResendConfirmationEndpoint: makeResendConfirmationEndpoint(ctx, svcFunc),
		CheckUsernameEndpoint:      makeCheckUsernameTakenEndpoint(ctx, svcFunc),
		LoginEndpoint:              makeLoginEndpoint(ctx, svcFunc),
	}
}

func makeRegisterEndpoint(ctx context.Context, svc grpcSvcConnFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cur userSvcEndpoint.CreateUserRequest

		err := json.NewDecoder(r.Body).Decode(&cur)
		if err != nil {
			http.Error(w, "Failed to decode json body", http.StatusInternalServerError)
			log.Printf("makeRegisterEndpoint: Failed to decode json body: %s\n", err)
			return
		}

		err = cur.Validate()
		if err != nil {
			http.Error(w, fmt.Sprintf("Payload is not valid: %s", err), http.StatusBadRequest)
			log.Printf("makeRegisterendpoint: Payload not valid: %s\n", err)
			return
		}

		userSvc, conn, err := svc()
		if err != nil {
			http.Error(w, "Failed to connect to service", http.StatusInternalServerError)
			log.Fatalf("makeRegisterEndpoint: Failed to connect to user grpc service: %s\n", err)
			return
		}
		defer conn.Close()

		_, err = userSvc.CreateUser(ctx, cur.Username, cur.Email, cur.Password)
		if err != nil {
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			log.Printf("makeRegisterEndpoint: Failed to create user when calling service: %s\n", err)
			return
		}

		cr := userSvcEndpoint.ConfirmResponse{Ok: true}
		jr, err := json.Marshal(cr)
		if err != nil {
			http.Error(w, "Failed to encode json", http.StatusInternalServerError)
			log.Printf("makeRegisterEndpoint: Failed marshal confirmResponse into json: %s\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jr)
	}
}

func makeConfirmUserEndpoint(ctx context.Context, svc grpcSvcConnFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cur userSvcEndpoint.ConfirmUserRequest

		err := json.NewDecoder(r.Body).Decode(&cur)
		if err != nil {
			http.Error(w, "Failed to decode json body", http.StatusInternalServerError)
			log.Fatalf("makeConfirmUserEndpoint: Failed to decode json body: %s\n", err)
			return
		}

		err = cur.Validate()
		if err != nil {
			http.Error(w, fmt.Sprintf("Payload is not valid: %s", err), http.StatusBadRequest)
			log.Printf("makeRegisterendpoint: Payload not valid: %s\n", err)
			return
		}

		userSvc, conn, err := svc()
		if err != nil {
			http.Error(w, "Failed to connect to service", http.StatusInternalServerError)
			log.Printf("makeConfirmUserEndpoint: Failed to connect to service: %s\n", err)
		}
		defer conn.Close()

		err = userSvc.ConfirmUser(ctx, cur.Username, cur.Code)
		if err != nil {
			http.Error(w, "Failed to confirm user", http.StatusInternalServerError)
			log.Printf("makeConfirmUserEndpoint: Failed to confirm user: %s\n", err)
			return
		}

		cr := userSvcEndpoint.ConfirmResponse{Ok: true}
		jr, err := json.Marshal(cr)
		if err != nil {
			http.Error(w, "Failed to encode json", http.StatusInternalServerError)
			log.Printf("makeConfirmUserEndpoint: Failed to marshal confirmResponse into json: %s\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jr)
	}
}

func makeResendConfirmationEndpoint(ctx context.Context, svc grpcSvcConnFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rcr userSvcEndpoint.ResendConfirmationRequest

		err := json.NewDecoder(r.Body).Decode(&rcr)
		if err != nil {
			http.Error(w, "Failed to decode json body", http.StatusInternalServerError)
			log.Printf("makeResendConfirmationEndpoint: Failed to decode json body: %s\n", err)
			return
		}

		err = rcr.Validate()
		if err != nil {
			http.Error(w, fmt.Sprintf("Payload is not valid: %s", err), http.StatusBadRequest)
			log.Printf("makeResendConfirmationEndpoint: Payload not valid: %s\n", err)
			return
		}

		userSvc, conn, err := svc()
		if err != nil {
			http.Error(w, "Failed to connect to service", http.StatusInternalServerError)
			log.Printf("makeResendConfirmationEndpoint: Failed to connect to service: %s\n", err)
			return
		}
		defer conn.Close()

		taken, err := userSvc.UsernameTaken(ctx, rcr.Username)
		if err != nil {
			http.Error(w, "Failed to check if username exists", http.StatusInternalServerError)
			log.Printf("makeResendConfirmationEndpoint: Failed to check if username exists: %s\n", err)
			return
		}

		if taken {
			err = userSvc.ResendConfirmation(ctx, rcr.Username)
			if err != nil {
				http.Error(w, "Failed to resend confirmation", http.StatusInternalServerError)
				log.Printf("makeResendConfirmationEndpoint: Failed to resend confirmation: %s\n", err)
				return
			}
		}

		cr := userSvcEndpoint.ConfirmResponse{Ok: taken}
		jr, err := json.Marshal(cr)
		if err != nil {
			http.Error(w, "Failed to encode json", http.StatusInternalServerError)
			log.Printf("makeResendConfirmationEndpoint: Failed to encode json: %s\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jr)
	}
}

func makeCheckUsernameTakenEndpoint(ctx context.Context, svc grpcSvcConnFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Required url param username missing", http.StatusInternalServerError)
			log.Printf("makeCheckUsernameTakenEndpoint: Required url param username missing\n")
			return
		}

		userSvc, conn, err := svc()
		if err != nil {
			http.Error(w, "Failed to connect to service", http.StatusInternalServerError)
			log.Printf("makeCheckUsernameTakenEndpoint: Failed to connect to service %s\n", err)
			return
		}
		defer conn.Close()

		taken, err := userSvc.UsernameTaken(ctx, username)
		if err != nil {
			http.Error(w, "Failed to check if username exists", http.StatusInternalServerError)
			log.Printf("makeCheckUsernameTakenEndpoint: Failed to check if username exists: %s\n", err)
			return
		}

		cr := userSvcEndpoint.ConfirmResponse{Ok: taken}
		jr, err := json.Marshal(cr)
		if err != nil {
			http.Error(w, "Failed to encode json", http.StatusInternalServerError)
			log.Printf("makeCheckUsernameTakenEndpoint: Failed to encode json: %s\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jr)
	}
}

func makeLoginEndpoint(ctx context.Context, svc grpcSvcConnFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var lr userSvcEndpoint.LoginRequest

		err := json.NewDecoder(r.Body).Decode(&lr)
		if err != nil {
			http.Error(w, "Failed to decode json body", http.StatusInternalServerError)
			log.Printf("makeLoginEndpoint: Failed to decode json: %s\n", err)
			return
		}

		err = lr.Validate()
		if err != nil {
			http.Error(w, fmt.Sprintf("Payload is not valid: %s", err), http.StatusBadRequest)
			log.Printf("makeLoginEndpoint: Payload not valid: %s\n", err)
			return
		}

		userSvc, conn, err := svc()
		if err != nil {
			http.Error(w, "Failed to connect to service", http.StatusInternalServerError)
			log.Printf("makeLoginEndpoint: Failed to connect to service %s\n", err)
			return
		}
		defer conn.Close()

		jwt, err := userSvc.Login(ctx, lr.Username, lr.Password)
		if err != nil {
			if strings.Contains(err.Error(), "UserNotConfirmedException") {
				http.Error(w, "User is not confirmed", http.StatusForbidden)
				log.Printf("makeLoginEndpoint: User is not confirmed: %s\n", err)
				return
			}

			http.Error(w, "Username or password is incorrect", http.StatusUnauthorized)
			log.Printf("makeLoginEndpoint: Failed to login user: %s\n", err)
			return
		}

		lgr := userSvcEndpoint.LoginResponse{Jwt: jwt}
		jr, err := json.Marshal(lgr)
		if err != nil {
			http.Error(w, "Failed to encode json", http.StatusInternalServerError)
			log.Printf("makeLoginEndpoint: Failed to marshal login response to json: %s\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jr)
	}
}
