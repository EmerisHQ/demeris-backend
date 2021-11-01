<template>
  <card-component title="Sign In" icon="lock">
    <form @submit.prevent="submit">
      <b-field horizontal label="Email">
        <b-input v-model="form.email" name="email" type="text" required />
      </b-field>
      <hr />
      <b-field horizontal label="Password">
        <b-input
          v-model="form.password"
          name="password"
          type="password"
          required
        />
      </b-field>
      <hr />
      <b-field horizontal>
        <div class="control">
          <button
            type="submit"
            class="button is-primary"
            :class="{ 'is-loading': isLoading }"
          >
            Login
          </button>
        </div>
      </b-field>
    </form>
  </card-component>
</template>

<script>
import CardComponent from "@/components/CardComponent";
export default {
  name: "SignInForm",
  components: {
    CardComponent,
  },
  data() {
    return {
      isLoading: false,
      form: {
        email: null,
        password: null,
      },
    };
  },
  methods: {
    submit() {
      let t = this;
      this.$fire.auth
        .signInWithEmailAndPassword(this.form.email, this.form.password)
        .catch(function (error) {
          this.$buefy.snackbar.open({
            type: 'is-danger',
            message: error.message,
            queue: false,
          });
        })
        .then((user) => {
          $nuxt.$router.push("/");
        });
    },
  },
};
</script>
